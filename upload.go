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
	"os"
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
	errc := make(chan error, 1)
	updc := make(chan *Upload)
	grp, ctx := errgroup.WithContext(ctx)

	uploadablesc, uperrc := uploadables.Uploadables(ctx)
	for i := 0; i < concurrency; i++ {
		grp.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case err := <-uperrc:
					return err
				case err := <-errc:
					return err
				case up, ok := <-uploadablesc:
					if !ok {
						log.Info().Msg("exiting; exhausted uploadables")
						return nil
					}
					log.Info().
						Str("name", up.Name).
						Str("album", up.AlbumID).
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
		})
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

// Uploadables is a factory for Uploadable instances
type Uploadables interface {
	// Uploadable returns an Uploadable instance or nil of no more Uploadable instances are available
	Uploadables(context.Context) (<-chan *Uploadable, <-chan error)
}

type fsUploadables struct {
	client    *Client
	albumID   string
	filenames []string
	config    *fsUploadablesConfig
}

type fsUploadablesConfig struct {
	extensions []string
}

type FsUploadablesOption func(c *fsUploadablesConfig)

func WithExtensions(exts ...string) FsUploadablesOption {
	return func(c *fsUploadablesConfig) {
		c.extensions = make([]string, len(exts))
		for i := range exts {
			c.extensions[i] = strings.ToLower(exts[i])
		}
	}
}

// NewFsUploadables returns a new instance of an Uploadables which creates Uploadable instances
//  from files on the filesystem
func NewFsUploadables(client *Client, albumID string, filenames []string, opts ...FsUploadablesOption) Uploadables {
	config := &fsUploadablesConfig{extensions: []string{".jpg"}}
	for i := range opts {
		opts[i](config)
	}
	return &fsUploadables{
		client:    client,
		albumID:   albumID,
		filenames: filenames,
		config:    config,
	}
}

func (p *fsUploadables) Uploadables(ctx context.Context) (<-chan *Uploadable, <-chan error) {
	errc := make(chan error, 1)
	images := make(map[string]*Image)

	log.Info().Msg("querying existing gallery images")
	if err := p.client.Image.ImagesIter(ctx, p.albumID, func(img *Image) (bool, error) {
		images[img.FileName] = img
		return true, nil
	}); err != nil {
		errc <- err
		return nil, errc
	}
	log.Info().Int("count", len(images)).Msg("existing gallery images")

	filenamesc := make(chan string)
	uploadablesc := make(chan *Uploadable)

	go func() {
		defer close(filenamesc)
		if err := p.walk(ctx, filenamesc); err != nil {
			errc <- err
		}
	}()

	go func() {
		defer close(errc)
		defer close(uploadablesc)
		for filename := range filenamesc {
			select {
			case <-ctx.Done():
				errc <- ctx.Err()
				return
			default:
				if !p.supported(filename) {
					log.Info().Str("path", filename).Msg("skipping; not supported")
					continue
				}
				up, err := p.uploadableFromFile(filename)
				if err != nil {
					errc <- err
					return
				}
				up.AlbumID = p.albumID
				img, ok := images[up.Name]
				if ok {
					if up.MD5 == img.ArchivedMD5 {
						log.Info().Str("path", filename).Msg("skipping; md5 matches")
						continue
					}
					up.Replaces = img.URIs.Image.URI
				}
				select {
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				case uploadablesc <- up:
					log.Info().Str("path", filename).Msg("uploadable")
				}
			}
		}
	}()

	return uploadablesc, errc
}

func (p *fsUploadables) uploadableFromFile(path string) (*Uploadable, error) {
	fp, err := os.Open(path)
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

func (p *fsUploadables) walk(ctx context.Context, filesc chan<- string) error {
	for _, root := range p.filenames {
		if err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
			if info.IsDir() {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case filesc <- path:
				log.Debug().Str("path", path).Msg("walk")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (p *fsUploadables) supported(filename string) bool {
	f := strings.ToLower(filename)
	for i := range p.config.extensions {
		if strings.HasSuffix(f, p.config.extensions[i]) {
			return true
		}
	}
	return false
}
