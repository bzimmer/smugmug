package smugmug

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// UploadService is the API for the upload endpoint
type UploadService service

// Upload an image to an album
func (s *UploadService) Upload(ctx context.Context, albumID string, up *Uploadable) (*Upload, error) {
	/*
		Documentation on the upload process is available at SmugMug

		https://api.smugmug.com/api/v2/doc/reference/upload.html
	*/

	uri := fmt.Sprintf("%s/%s/%s", uploadURL, albumID, up.Name)
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
		"X-Smug-AlbumUri":     "/api/v2/album/" + albumID,
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

func (s *UploadService) Uploader(ctx context.Context, albumID string) (*Uploader, error) {
	up := &Uploader{
		client:  s.client,
		albumID: albumID,
		images:  make(map[string]*Image),
	}
	if err := s.client.Image.ImagesIter(ctx, albumID, func(img *Image) (bool, error) {
		up.images[img.FileName] = img
		return true, nil
	}); err != nil {
		return nil, err
	}
	return up, nil
}

type Uploads struct {
	Upload *Upload
	Err    error
}

type Uploader struct {
	client  *Client
	albumID string
	uploads []*Uploadable
	images  map[string]*Image
}

func (u *Uploader) Add(uploads ...*Uploadable) {
	u.uploads = append(u.uploads, uploads...)
}

func (u *Uploader) Upload(ctx context.Context) <-chan *Uploads {
	res := make(chan *Uploads)
	go func() {
		defer close(res)
		// @todo upload
	}()
	return res
}

func UploadableFromFile(path string) (*Uploadable, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	buf := bytes.NewBuffer(nil)
	n, err := io.Copy(buf, fp)
	if err != nil {
		return nil, err
	}
	hash := md5.Sum(buf.Bytes())

	return &Uploadable{
		Name:   filepath.Base(path),
		Size:   n,
		MD5:    fmt.Sprintf("%x", hash),
		Reader: bytes.NewBuffer(buf.Bytes()),
	}, nil
}
