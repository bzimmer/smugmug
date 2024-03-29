package filesystem

import (
	"bytes"
	"crypto/md5" //nolint:gosec // used to match md5 at smugmug
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/bzimmer/smugmug"
)

// PreFunc is called before the file is opened
type PreFunc func(fs afero.Fs, filename string) (bool, error)

// UseFunc is called after the file is opened but before being uploaded
type UseFunc func(up *smugmug.Uploadable) error

// FsUploadable creates Uploadable instances
type FsUploadable interface {
	// Uploadable creates an Uploadable from a filesystem
	Uploadable(afero.Fs, string) (*smugmug.Uploadable, error)
	// Pre registers a PreFunc
	Pre(...PreFunc)
	// Use registers a UseFunc
	Use(...UseFunc)
}

// ErrSkip is used to skip an Uploadable
var ErrSkip = errors.New("skip")

// Extensions represents the valid list of extensions to upload
func Extensions(extension ...string) PreFunc {
	return func(_ afero.Fs, filename string) (bool, error) {
		f := strings.ToLower(filename)
		for i := range extension {
			if strings.HasSuffix(f, extension[i]) {
				return true, nil
			}
		}
		return false, nil
	}
}

// Skip checks if the Uploadable is already uploaded by comparing MD5s
// If `force` is true the Uploadable will be always be uploaded
func Skip(force bool, images map[string]*smugmug.Image) UseFunc {
	return func(up *smugmug.Uploadable) error {
		if force {
			return nil
		}
		img, ok := images[up.Name]
		if !ok {
			return nil
		}
		if up.MD5 == img.ArchivedMD5 {
			return ErrSkip
		}
		return nil
	}
}

// Replace will update the Uploadable's URI if the image was already uploaded
// If `update` is false the URI will not be updated
func Replace(update bool, images map[string]*smugmug.Image) UseFunc {
	return func(up *smugmug.Uploadable) error {
		if !update {
			return nil
		}
		img, ok := images[up.Name]
		if ok {
			up.Replaces = img.URIs.Image.URI
		}
		return nil
	}
}

type fsUploadable struct {
	albumKey string
	pre      []PreFunc
	use      []UseFunc
}

// NewFsUploadable returns a newly instantiated FsUploadable instance
func NewFsUploadable(albumKey string) (FsUploadable, error) {
	if albumKey == "" {
		return nil, errors.New("missing albumKey")
	}
	return &fsUploadable{albumKey: albumKey}, nil
}

func (p *fsUploadable) Pre(f ...PreFunc) {
	p.pre = append(p.pre, f...)
}

func (p *fsUploadable) Use(f ...UseFunc) {
	p.use = append(p.use, f...)
}

func (p *fsUploadable) Uploadable(fs afero.Fs, filename string) (*smugmug.Uploadable, error) {
	for i := range p.pre {
		ok, err := p.pre[i](fs, filename)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrSkip
		}
	}

	up, err := p.open(fs, filename)
	if err != nil {
		return nil, err
	}
	up.AlbumKey = p.albumKey

	for i := range p.use {
		err = p.use[i](up)
		if err != nil {
			return nil, err
		}
	}

	return up, nil
}

func (p *fsUploadable) open(fs afero.Fs, path string) (*smugmug.Uploadable, error) {
	fp, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	var buf bytes.Buffer
	size, err := io.Copy(&buf, fp)
	if err != nil {
		return nil, err
	}

	return &smugmug.Uploadable{
		Name:   filepath.Base(path),
		Size:   size,
		MD5:    fmt.Sprintf("%x", md5.Sum(buf.Bytes())), //nolint:gosec // required for smugmug
		Reader: &buf,
	}, nil
}
