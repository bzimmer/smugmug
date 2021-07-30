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

func (s *UploadService) Uploads(ctx context.Context, provider Uploadables) (<-chan *Upload, <-chan error) {
	errc := make(chan error, 1)
	if err := provider.Begin(ctx); err != nil {
		errc <- err
		close(errc)
		return nil, errc
	}

	updc := make(chan *Upload)
	grp, ctx := errgroup.WithContext(ctx)
	for i := 0; i < concurrency; i++ {
		grp.Go(func() error {
			for {
				up, err := provider.Uploadable(ctx)
				if err != nil {
					return err
				}
				if up == nil {
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
	// Begin is called before any Uploadable is uploaded
	Begin(context.Context) error
	// Uploadable returns an Uploadable instance or nil of no more Uploadable instances are available
	Uploadable(context.Context) (*Uploadable, error)
}

type fsUploadables struct {
	client    *Client
	albumID   string
	images    map[string]*Image
	filenames []string
	errc      chan error
	filesc    chan string
}

// NewFsUploadables returns a new instance of an Uploadables which creates Uploadable instances
//  from files on the filesystem
func NewFsUploadables(client *Client, albumID string, filenames []string) Uploadables {
	return &fsUploadables{
		client:    client,
		albumID:   albumID,
		images:    make(map[string]*Image),
		filenames: filenames,
		filesc:    make(chan string),
		errc:      make(chan error, 1),
	}
}

func (p *fsUploadables) Begin(ctx context.Context) error {
	go func() {
		defer close(p.errc)
		defer close(p.filesc)
		n, err := p.walk(ctx)
		if err != nil {
			log.Error().Err(err).Int("enqueued", n).Msg("walk")
		} else {
			log.Info().Int("enqueued", n).Msg("walk")
		}
	}()
	if err := p.client.Image.ImagesIter(ctx, p.albumID, func(img *Image) (bool, error) {
		p.images[img.FileName] = img
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

func (p *fsUploadables) Uploadable(ctx context.Context) (*Uploadable, error) {
	for filename := range p.filesc {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			up, err := p.uploadableFromFile(filename)
			if err != nil {
				return nil, err
			}
			up.AlbumID = p.albumID
			img, ok := p.images[up.Name]
			if ok {
				if up.MD5 == img.ArchivedMD5 {
					log.Info().Str("path", filename).Msg("skipping; md5 matches")
					continue
				}
				up.Replaces = img.URIs.Image.URI
			}
			return up, nil
		}
	}
	return nil, nil
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

func (p *fsUploadables) walk(ctx context.Context) (int, error) {
	var n int
	for _, filename := range p.filenames {
		if err := filepath.Walk(filename, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if supported(path) {
				n++
				p.filesc <- path
			}
			return nil
		}); err != nil {
			select {
			case <-ctx.Done():
				return n, ctx.Err()
			case p.errc <- err:
				return n, err
			}
		}
	}
	return n, nil
}

func supported(filename string) bool {
	return strings.HasSuffix(filename, ".jpg")
}
