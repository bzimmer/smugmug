package filesystem

import (
	"context"
	"io/fs"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/bzimmer/smugmug"
)

type fsUploadables struct {
	fsys       fs.FS
	filenames  []string
	uploadable FsUploadable
}

// NewFsUploadables returns a new instance of an Uploadables which creates Uploadable instances
//  from files on the filesystem
func NewFsUploadables(fsys fs.FS, filenames []string, uploadable FsUploadable) smugmug.Uploadables {
	return &fsUploadables{fsys: fsys, filenames: filenames, uploadable: uploadable}
}

func (p *fsUploadables) Uploadables(ctx context.Context) (<-chan *smugmug.Uploadable, <-chan error) {
	grp, ctx := errgroup.WithContext(ctx)

	errc := make(chan error, 1)
	filenamesc, walkerrc := p.walk(ctx)
	grp.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-walkerrc:
			return err
		}
	})

	uploadablesc := make(chan *smugmug.Uploadable)
	grp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case filename, ok := <-filenamesc:
				if !ok {
					return nil
				}
				up, err := p.uploadable.Uploadable(p.fsys, filename)
				if err != nil {
					return err
				}
				if up == nil {
					continue
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case uploadablesc <- up:
					log.Info().Str("path", filename).Msg("uploadable")
				}
			}
		}
	})

	go func() {
		defer close(errc)
		defer close(uploadablesc)
		if err := grp.Wait(); err != nil {
			errc <- err
		}
	}()

	return uploadablesc, errc
}

func (p *fsUploadables) walk(ctx context.Context) (<-chan string, <-chan error) {
	errc := make(chan error, 1)
	filenamesc := make(chan string)
	go func() {
		defer close(errc)
		defer close(filenamesc)
		for _, root := range p.filenames {
			if err := fs.WalkDir(p.fsys, root, func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case filenamesc <- path:
					log.Debug().Str("path", path).Msg("walk")
				}
				return nil
			}); err != nil {
				errc <- err
			}
		}
	}()
	return filenamesc, errc
}
