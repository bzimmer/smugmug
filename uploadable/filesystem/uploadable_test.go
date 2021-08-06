package filesystem_test

import (
	"testing"

	"github.com/armon/go-metrics"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
)

func TestUploadable(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	albumKey := "Du82xY"

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

	for i := range tests {
		test := tests[i]
		t.Run(test.filename, func(t *testing.T) {
			fs := new(afero.MemMapFs)
			a.NoError(afero.WriteFile(fs, test.filename, []byte("this is a test"), 0644))

			mt := &metrics.Metrics{}
			options := append(test.options,
				filesystem.WithImages(albumKey, test.images),
				filesystem.WithMetrics(mt))
			fsup, err := filesystem.NewFsUploadable(options...)
			a.NoError(err)

			up, err := fsup.Uploadable(fs, test.filename)
			a.NoError(err)
			switch test.none {
			case true:
				a.Nil(up)
			case false:
				a.NotNil(up)
			}
		})
	}
}
