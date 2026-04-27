package filesystem_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
)

func TestNewUploadable(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	u, err := filesystem.NewFsUploadable("")
	a.Error(err)
	a.Nil(u)
}

func TestUploadable(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	errPre := errors.New("pre error")

	tests := []struct {
		name     string
		filename string
		pre      filesystem.PreFunc
		use      filesystem.UseFunc
		f        func(*smugmug.Uploadable, error)
	}{
		{
			name:     "extensions filter skips non-matching file",
			filename: ".DS_Info",
			pre:      filesystem.Extensions(".jpg"),
			f: func(up *smugmug.Uploadable, err error) {
				a.Error(err)
				a.ErrorIs(err, filesystem.ErrSkip)
				a.Nil(up)
			},
		},
		{
			name:     "extensions filter passes matching file",
			filename: "DSC1234.jpg",
			pre:      filesystem.Extensions(".jpg"),
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
			},
		},
		{
			name:     "no pre func passes file",
			filename: "DSC1234.jpg",
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
			},
		},
		{
			name:     "skip by md5 match",
			filename: "DSC12345.jpg",
			use: filesystem.Skip(false, map[string]*smugmug.Image{
				"DSC12345.jpg": {ArchivedMD5: "54b0c58c7ce9f2a8b551351102ee0938"},
			}),
			f: func(up *smugmug.Uploadable, err error) {
				a.Error(err)
				a.ErrorIs(err, filesystem.ErrSkip)
				a.Nil(up)
			},
		},
		{
			name:     "skip returns nil when force is true",
			filename: "DSC12345.jpg",
			use: filesystem.Skip(true, map[string]*smugmug.Image{
				"DSC12345.jpg": {ArchivedMD5: "54b0c58c7ce9f2a8b551351102ee0938"},
			}),
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
			},
		},
		{
			name:     "skip returns nil when image not in map",
			filename: "DSC99999.jpg",
			use: filesystem.Skip(false, map[string]*smugmug.Image{
				"other.jpg": {ArchivedMD5: "abc123"},
			}),
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
			},
		},
		{
			name:     "skip returns nil when md5 differs",
			filename: "DSC12345.jpg",
			use: filesystem.Skip(false, map[string]*smugmug.Image{
				"DSC12345.jpg": {ArchivedMD5: "different_md5"},
			}),
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
			},
		},
		{
			name:     "replace sets Replaces URI",
			filename: "DSC12345.jpg",
			use: filesystem.Replace(true, map[string]*smugmug.Image{
				"DSC12345.jpg": {URIs: smugmug.ImageURIs{Image: &smugmug.APIEndpoint{URI: "foo"}}},
			}),
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
				a.Equal("foo", up.Replaces)
			},
		},
		{
			name:     "replace does not set Replaces URI when update=false",
			filename: "DSC12345.jpg",
			use: filesystem.Replace(false, map[string]*smugmug.Image{
				"DSC12345.jpg": {URIs: smugmug.ImageURIs{Image: &smugmug.APIEndpoint{URI: "foo"}}},
			}),
			f: func(up *smugmug.Uploadable, err error) {
				a.NoError(err)
				a.NotNil(up)
				a.Empty(up.Replaces)
			},
		},
		{
			name:     "pre function returning error propagates error",
			filename: "DSC_error.jpg",
			pre: func(_ afero.Fs, _ string) (bool, error) {
				return false, errPre
			},
			f: func(up *smugmug.Uploadable, err error) {
				a.Error(err)
				a.ErrorIs(err, errPre)
				a.Nil(up)
			},
		},
	}

	albumKey := "Du82xY"
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fs := new(afero.MemMapFs)
			a.NoError(afero.WriteFile(fs, test.filename, []byte("this is a test"), 0644))
			fsup, err := filesystem.NewFsUploadable(albumKey)
			a.NoError(err)
			if test.pre != nil {
				fsup.Pre(test.pre)
			}
			if test.use != nil {
				fsup.Use(test.use)
			}
			up, err := fsup.Uploadable(fs, test.filename)
			test.f(up, err)
		})
	}
}

func TestOpenError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// Use an empty MemMapFs so the file does not exist.
	fs := new(afero.MemMapFs)
	fsup, err := filesystem.NewFsUploadable("albumKey")
	a.NoError(err)

	// No pre function so we reach open() directly.
	up, err := fsup.Uploadable(fs, "nonexistent.jpg")
	a.Error(err)
	a.Nil(up)
}
