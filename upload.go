package smugmug

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// UploadService is the API for the upload endpoint
type UploadService service

const concurrency = 5

// Upload an image to an album
func (s *UploadService) Upload(ctx context.Context, up *Uploadable) (*Upload, error) {
	/*
		Documentation on the upload process is available at SmugMug

		https://api.smugmug.com/api/v2/doc/reference/upload.html
	*/

	if up.AlbumID == "" {
		return nil, errors.New("missing albumID")
	}

	uri := fmt.Sprintf("%s/%s", uploadURL, up.Name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, up.Reader)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Accept":              "application/json",
		"Content-MD5":         up.MD5,
		"Content-Length":      strconv.FormatInt(up.Size, 10),
		"User-Agent":          userAgent,
		"X-Smug-Version":      "v2",
		"X-Smug-AlbumUri":     "/api/v2/album/" + up.AlbumID,
		"X-Smug-ResponseType": "JSON",
	}

	if up.Replaces != "" {
		headers["X-Smug-ImageUri"] = up.Replaces
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	res := &Upload{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *UploadService) Uploads(ctx context.Context, uploadables Uploadables) (<-chan *Upload, <-chan error) {
	updc := make(chan *Upload)
	errc := make(chan error, 1)
	grp, ctx := errgroup.WithContext(ctx)

	uploadablesc, uperrc := uploadables.Uploadables(ctx)
	grp.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-uperrc:
			return err
		}
	})
	for i := 0; i < concurrency; i++ {
		grp.Go(s.uploads(ctx, uploadablesc, updc))
	}

	go func() {
		defer close(errc)
		defer close(updc)
		if err := grp.Wait(); err != nil {
			errc <- err
		}
	}()

	return updc, errc
}

func (s *UploadService) uploads(ctx context.Context,
	uploadablesc <-chan *Uploadable, updc chan<- *Upload) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case up, ok := <-uploadablesc:
				if !ok {
					log.Debug().Msg("exiting; exhausted uploadables")
					return nil
				}
				if up == nil {
					continue
				}
				log.Info().
					Str("name", up.Name).
					Str("album", up.AlbumID).
					Str("replaces", up.Replaces).
					Msg("uploading")
				upload, err := s.Upload(ctx, up)
				if err != nil {
					log.Error().
						Err(err).
						Str("name", up.Name).
						Str("album", up.AlbumID).
						Msg("failed")
					return err
				}
				log.Info().
					Str("name", up.Name).
					Str("album", up.AlbumID).
					Str("uri", upload.UploadedImage.ImageURI).
					Msg("uploaded")
				select {
				case <-ctx.Done():
					return ctx.Err()
				case updc <- upload:
				}
			}
		}
	}
}

// Uploadables is a factory for Uploadable instances from a filesystem
type Uploadables interface {
	// Uploadable returns an Uploadable instance or nil of no more Uploadable instances are available
	Uploadables(context.Context) (<-chan *Uploadable, <-chan error)
}

// FsUploadable creates Uploadable instances
type FsUploadable interface {
	// Uploadable creates an Uploadable from a filesystem
	Uploadable(fs.FS, string) (*Uploadable, error)
}

type fsUploadable struct {
	skip       bool
	replace    bool
	albumID    string
	extensions []string
	images     map[string]*Image
}

type FsUploadableOption func(c *fsUploadable)

// WithExtensions configures which extensions (inclusive of '.') are supported
func WithExtensions(exts ...string) FsUploadableOption {
	return func(c *fsUploadable) {
		c.extensions = make([]string, len(exts))
		for i := range exts {
			c.extensions[i] = strings.ToLower(exts[i])
		}
	}
}

// WithReplace configures if an image replaces an image with the same filename or creates a duplicate
func WithReplace(replace bool) FsUploadableOption {
	return func(c *fsUploadable) {
		c.replace = replace
	}
}

// WithSkip configures if an image is uploaded if the md5 sum matches
func WithSkip(skip bool) FsUploadableOption {
	return func(c *fsUploadable) {
		c.skip = skip
	}
}

// WithImages maps basenames from the album to existing images
func WithImages(albumID string, images map[string]*Image) FsUploadableOption {
	return func(c *fsUploadable) {
		c.images = images
		c.albumID = albumID
	}
}

func NewFsUploadable(options ...FsUploadableOption) (FsUploadable, error) {
	p := &fsUploadable{skip: true, replace: true}
	for i := range options {
		options[i](p)
	}
	if p.images == nil {
		p.images = make(map[string]*Image)
	}
	if p.albumID == "" {
		return nil, errors.New("missing albumID")
	}
	return p, nil
}

func (p *fsUploadable) Uploadable(fsys fs.FS, filename string) (*Uploadable, error) {
	if !p.supported(filename) {
		log.Info().Str("reason", "unsupported").Str("path", filename).Msg("skipping")
		return nil, nil
	}
	up, err := p.open(fsys, filename)
	if err != nil {
		return nil, err
	}
	up.AlbumID = p.albumID
	img, ok := p.images[up.Name]
	if ok {
		if p.skip && up.MD5 == img.ArchivedMD5 {
			log.Info().Str("reason", "md5").Str("path", filename).Msg("skipping")
			return nil, nil
		}
		if p.replace && img.URIs.Image != nil {
			up.Replaces = img.URIs.Image.URI
		}
	}
	return up, nil
}

func (p *fsUploadable) open(fsys fs.FS, path string) (*Uploadable, error) {
	fp, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	buf := bytes.NewBuffer(nil)
	size, err := io.Copy(buf, fp)
	if err != nil {
		return nil, err
	}

	return &Uploadable{
		Name:   filepath.Base(path),
		Size:   size,
		MD5:    fmt.Sprintf("%x", md5.Sum(buf.Bytes())),
		Reader: bytes.NewBuffer(buf.Bytes()),
	}, nil
}

func (p *fsUploadable) supported(filename string) bool {
	f := strings.ToLower(filename)
	for i := range p.extensions {
		if strings.HasSuffix(f, p.extensions[i]) {
			return true
		}
	}
	return false
}

type fsUploadables struct {
	fsys       fs.FS
	filenames  []string
	uploadable FsUploadable
}

// NewFsUploadables returns a new instance of an Uploadables which creates Uploadable instances
//  from files on the filesystem
func NewFsUploadables(fsys fs.FS, filenames []string, uploadable FsUploadable) Uploadables {
	return &fsUploadables{fsys: fsys, filenames: filenames, uploadable: uploadable}
}

func (p *fsUploadables) Uploadables(ctx context.Context) (<-chan *Uploadable, <-chan error) {
	grp, ctx := errgroup.WithContext(ctx)

	errc := make(chan error, 1)
	filenamesc, walkerrc := p.walk(ctx)
	grp.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-walkerrc:
			return err
		}
	})

	uploadablesc := make(chan *Uploadable)
	grp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case filename, ok := <-filenamesc:
				if !ok {
					return nil
				}
				up, err := p.uploadable.Uploadable(p.fsys, filename)
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case uploadablesc <- up:
					log.Info().Str("path", filename).Msg("uploadable")
				}
			}
		}
	})

	go func() {
		defer close(errc)
		defer close(uploadablesc)
		if err := grp.Wait(); err != nil {
			errc <- err
		}
	}()

	return uploadablesc, errc
}

func (p *fsUploadables) walk(ctx context.Context) (<-chan string, <-chan error) {
	errc := make(chan error, 1)
	filenamesc := make(chan string)
	go func() {
		defer close(errc)
		defer close(filenamesc)
		for _, root := range p.filenames {
			if err := fs.WalkDir(p.fsys, root, func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case filenamesc <- path:
					log.Debug().Str("path", path).Msg("walk")
				}
				return nil
			}); err != nil {
				fmt.Println(err)
				errc <- err
			}
		}
	}()
	return filenamesc, errc
}
