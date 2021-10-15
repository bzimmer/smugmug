package filesystem

import (
	"bytes"
	"crypto/md5" //nolint:gosec
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

// UseFunc is called after the file is opened but before it sent to be uploaded
type UseFunc func(up *smugmug.Uploadable) (*smugmug.Uploadable, error)

// FsUploadable creates Uploadable instances
type FsUploadable interface {
	// Uploadable creates an Uploadable from a filesystem
	Uploadable(afero.Fs, string) (*smugmug.Uploadable, error)
	// Pre registers a PreFunc
	Pre(...PreFunc)
	// Use registers a UseFunc
	Use(...UseFunc)
}

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
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		if force {
			return up, nil
		}
		img, ok := images[up.Name]
		if !ok {
			return up, nil
		}
		if up.MD5 == img.ArchivedMD5 {
			return nil, nil
		}
		return up, nil
	}
}

// Replace will update the Uploadable's URI if the image was already uploaded
// If `update` is false the URI will not be updated
func Replace(update bool, images map[string]*smugmug.Image) UseFunc {
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		if !update {
			return up, nil
		}
		img, ok := images[up.Name]
		if ok {
			up.Replaces = img.URIs.Image.URI
		}
		return up, nil
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
			return nil, nil
		}
	}

	up, err := p.open(fs, filename)
	if err != nil {
		return nil, err
	}
	up.AlbumKey = p.albumKey

	for i := range p.use {
		up, err = p.use[i](up)
		if err != nil {
			return nil, err
		}
		if up == nil {
			return nil, nil
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

	buf := bytes.NewBuffer(nil)
	size, err := io.Copy(buf, fp)
	if err != nil {
		return nil, err
	}

	return &smugmug.Uploadable{
		Name:   filepath.Base(path),
		Size:   size,
		MD5:    fmt.Sprintf("%x", md5.Sum(buf.Bytes())), //nolint:gosec
		Reader: bytes.NewBuffer(buf.Bytes()),
	}, nil
}
