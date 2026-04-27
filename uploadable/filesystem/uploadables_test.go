package filesystem_test

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
)

type testFsUploadable struct{}

func (t *testFsUploadable) Uploadable(_ afero.Fs, filename string) (*smugmug.Uploadable, error) {
	switch filename {
	case "DSC4321.jpg":
		return &smugmug.Uploadable{
			Name: "DSC4321.jpg",
		}, nil
	case "Readme.md":
		return nil, filesystem.ErrSkip
	default:
		return nil, errors.New("missing")
	}
}

func (t *testFsUploadable) Pre(_ ...filesystem.PreFunc) {}
func (t *testFsUploadable) Use(_ ...filesystem.UseFunc) {}

func TestUploadables(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	fs := new(afero.MemMapFs)
	a.NoError(fs.MkdirAll("Directory", 0755))
	a.NoError(afero.WriteFile(fs, "DSC4321.jpg", []byte("this is a test"), 0644))
	a.NoError(afero.WriteFile(fs, "Readme.md", []byte("bah"), 0644))

	filenames := []string{"DSC4321.jpg", "Readme.md", "Directory", "missing.txt"}
	fu := filesystem.NewFsUploadables(fs, filenames, &testFsUploadable{})
	a.NotNil(fu)
	upc, errc := fu.Uploadables(context.TODO())

	up := <-upc
	a.NotNil(up)
	a.Equal("DSC4321.jpg", up.Name)
	up = <-upc
	a.Nil(up)
	err := <-errc
	a.Error(err)
}

func TestUploadablesNonSkipError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// testFsUploadable returns errors.New("missing") for "missing.txt",
	// which is not ErrSkip, so it propagates as an error.
	fs := new(afero.MemMapFs)
	a.NoError(afero.WriteFile(fs, "missing.txt", []byte("content"), 0644))

	fu := filesystem.NewFsUploadables(fs, []string{"missing.txt"}, &testFsUploadable{})
	a.NotNil(fu)
	_, errc := fu.Uploadables(context.TODO())

	err := <-errc
	a.Error(err)
	a.Equal("missing", err.Error())
}
