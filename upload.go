package smugmug

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// UploadService is the API for the upload endpoint
type UploadService service

// Uploadables is a factory for Uploadable instances
type Uploadables interface {
	// Uploadables returns a channel of Uploadable instances
	Uploadables(context.Context) (<-chan *Uploadable, <-chan error)
}

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

	uri := fmt.Sprintf("%s/%s", s.client.uploadURL, up.Name)
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
