package filesystem

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/armon/go-metrics"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"

	"github.com/bzimmer/smugmug"
)

// FsUploadable creates Uploadable instances
type FsUploadable interface {
	// Uploadable creates an Uploadable from a filesystem
	Uploadable(afero.Fs, string) (*smugmug.Uploadable, error)
}

type fsUploadable struct {
	skip       bool
	replace    bool
	albumKey   string
	extensions []string
	metrics    *metrics.Metrics
	images     map[string]*smugmug.Image
}

// FsUploadableOption enables configuration of uploadable creation
type FsUploadableOption func(c *fsUploadable)

// WithMetrics configures the metrics instance
func WithMetrics(metrics *metrics.Metrics) FsUploadableOption {
	return func(c *fsUploadable) {
		c.metrics = metrics
	}
}

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
func WithImages(albumKey string, images map[string]*smugmug.Image) FsUploadableOption {
	return func(c *fsUploadable) {
		c.images = images
		c.albumKey = albumKey
	}
}

// NewFsUploadable returns a newly instantiated FsUploadable instance
func NewFsUploadable(options ...FsUploadableOption) (FsUploadable, error) {
	p := &fsUploadable{skip: true, replace: true}
	for i := range options {
		options[i](p)
	}
	if p.images == nil {
		p.images = make(map[string]*smugmug.Image)
	}
	if p.albumKey == "" {
		return nil, errors.New("missing albumKey")
	}
	if p.metrics == nil {
		p.metrics = metrics.Default()
	}
	return p, nil
}

func (p *fsUploadable) Uploadable(fs afero.Fs, filename string) (*smugmug.Uploadable, error) {
	if !p.supported(filename) {
		p.metrics.IncrCounter([]string{"fsUploadable", "skip", "unsupported"}, 1)
		log.Info().Str("reason", "unsupported").Str("path", filename).Msg("skipping")
		return nil, nil
	}
	up, err := p.open(fs, filename)
	if err != nil {
		return nil, err
	}
	up.AlbumKey = p.albumKey
	img, ok := p.images[up.Name]
	if ok {
		if p.skip && up.MD5 == img.ArchivedMD5 {
			p.metrics.IncrCounter([]string{"fsUploadable", "skip", "md5"}, 1)
			log.Info().Str("reason", "md5").Str("path", filename).Msg("skipping")
			return nil, nil
		}
		if p.replace && img.URIs.Image != nil {
			up.Replaces = img.URIs.Image.URI
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

	p.metrics.IncrCounter([]string{"fsUploadable", "open"}, 1)

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
