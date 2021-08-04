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

	tests := []string{
		"fs_test.go",
		"./fs_test.go",
		filepath.Join(dir, "fs_test.go"),
	}

	fs, err := filesystem.RelativeFS(dir)
	a.NoError(err)
	for i := range tests {
		test := tests[i]
		fp, err := fs.Open(test)
		a.NoError(err)
		a.NoError(fp.Close())
	}
}
