package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

type relativeFS struct {
	fs.FS
	root string
}

func (rfs *relativeFS) Open(name string) (fs.File, error) {
	name, err := filepath.Rel(rfs.root, name)
	if err != nil {
		return nil, err
	}
	return rfs.FS.Open(name)
}

// RelativeFS returns an fs.FS instance relative to `dir`
func RelativeFS(dir string) fs.FS {
	return &relativeFS{FS: os.DirFS(dir), root: dir}
}
