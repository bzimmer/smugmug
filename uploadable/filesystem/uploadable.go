package filesystem

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/bzimmer/smugmug"
)

// FsUploadable creates Uploadable instances
type FsUploadable interface {
	// Uploadable creates an Uploadable from a filesystem
	Uploadable(fs.FS, string) (*smugmug.Uploadable, error)
}

type fsUploadable struct {
	skip       bool
	replace    bool
	albumID    string
	extensions []string
	images     map[string]*smugmug.Image
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
func WithImages(albumID string, images map[string]*smugmug.Image) FsUploadableOption {
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
		p.images = make(map[string]*smugmug.Image)
	}
	if p.albumID == "" {
		return nil, errors.New("missing albumID")
	}
	return p, nil
}

func (p *fsUploadable) Uploadable(fsys fs.FS, filename string) (*smugmug.Uploadable, error) {
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

func (p *fsUploadable) open(fsys fs.FS, path string) (*smugmug.Uploadable, error) {
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

	return &smugmug.Uploadable{
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
