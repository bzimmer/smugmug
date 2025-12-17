package filesystem_test

import (
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

	tests := []struct {
		filename string
		replace  string
		skip     bool
		pre      filesystem.PreFunc
		use      filesystem.UseFunc
		uri      string
	}{
		{
			filename: ".DS_Info",
			skip:     true,
			pre:      filesystem.Extensions(".jpg"),
		},
		{
			filename: "DSC1234.jpg",
			skip:     false,
		},
		{
			filename: "DSC12345.jpg",
			skip:     true,
			use: filesystem.Skip(false, map[string]*smugmug.Image{
				"DSC12345.jpg": {
					ArchivedMD5: "54b0c58c7ce9f2a8b551351102ee0938",
				},
			}),
		},
		{
			filename: "DSC12345.jpg",
			skip:     false,
			uri:      "foo",
			use: filesystem.Replace(true, map[string]*smugmug.Image{
				"DSC12345.jpg": {
					URIs: smugmug.ImageURIs{
						Image: &smugmug.APIEndpoint{
							URI: "foo",
						},
					},
				},
			}),
		},
		{
			filename: "DSC12345.jpg",
			skip:     false,
			uri:      "",
			use: filesystem.Replace(false, map[string]*smugmug.Image{
				"DSC12345.jpg": {
					URIs: smugmug.ImageURIs{
						Image: &smugmug.APIEndpoint{
							URI: "foo",
						},
					},
				},
			}),
		},
	}

	albumKey := "Du82xY"
	for i := range tests {
		test := tests[i]
		t.Run(test.filename, func(t *testing.T) {
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
			switch test.skip {
			case true:
				a.Error(err)
				a.ErrorIs(err, filesystem.ErrSkip)
				a.Nil(up)
			case false:
				a.NotNil(up)
				if test.uri != "" {
					a.Equal(test.uri, up.Replaces)
				}
				if test.uri == "" {
					a.Empty(up.Replaces)
				}
			}
		})
	}
}
