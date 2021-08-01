package filesystem_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
)

func TestUploadable(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	albumID := "Du82xY"

	tests := []struct {
		filename string
		options  []filesystem.FsUploadableOption
		images   map[string]*smugmug.Image
		replace  string
		none     bool
	}{
		{filename: ".DS_Info", none: true},
		{filename: "DSC1234.jpg", none: true},
		{filename: "DSC1234.jpg", none: false, options: []filesystem.FsUploadableOption{
			filesystem.WithExtensions(".jpg"),
		}},
		{filename: "DSC12345.jpg", none: false,
			images: map[string]*smugmug.Image{
				"DSC12345.jpg": {
					ArchivedMD5: "e19c1283c925b3206685ff522acfe3e6",
				},
			},
			options: []filesystem.FsUploadableOption{
				filesystem.WithExtensions(".jpg"),
				filesystem.WithReplace(true),
				filesystem.WithSkip(true)},
		},
	}

	for _, tt := range tests {
		tt := tt

		options := append(tt.options, filesystem.WithImages(albumID, tt.images))
		fsup, err := filesystem.NewFsUploadable(options...)
		a.NoError(err)

		fsys := fstest.MapFS{
			tt.filename: {Data: []byte("this is a test")},
		}

		up, err := fsup.Uploadable(fsys, tt.filename)
		a.NoError(err)
		switch tt.none {
		case true:
			a.Nil(up)
		case false:
			a.NotNil(up)
		}
	}
}
