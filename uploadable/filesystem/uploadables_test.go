package filesystem_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
)

type testFsUploadable struct{}

func (t *testFsUploadable) Uploadable(fsys fs.FS, filename string) (*smugmug.Uploadable, error) {
	switch filename {
	case "DSC4321.jpg":
		return &smugmug.Uploadable{
			Name: "DSC4321.jpg",
		}, nil
	case "Readme.md":
		return nil, nil
	default:
		return nil, errors.New("missing")
	}
}

func TestUploadables(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	fsys := fstest.MapFS{
		"DSC4321.jpg": {Data: []byte("this is a test")},
		"Readme.md":   {Data: []byte("bah")},
		"Directory":   {Mode: fs.ModeDir},
	}

	filenames := []string{"DSC4321.jpg", "Readme.md", "Directory", "missing.txt"}
	fu := filesystem.NewFsUploadables(fsys, filenames, &testFsUploadable{})
	a.NotNil(fu)
	upc, errc := fu.Uploadables(context.Background())

	up := <-upc
	a.NotNil(up)
	a.Equal("DSC4321.jpg", up.Name)
	up = <-upc
	a.Nil(up)
	err := <-errc
	a.Error(err)
}
