package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestRelativeFS(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	dir, err := os.Getwd()
	a.NoError(err)

	fs := filesystem.RelativeFS(dir)
	fp, err := fs.Open(filepath.Join(dir, "fs_test.go"))
	a.NoError(err)
	defer fp.Close()
}
