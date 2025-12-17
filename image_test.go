package smugmug_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
)

func TestImage(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name       string
		imageKey   string
		expansions []string
		filename   string
		options    []smugmug.APIOption
		patch      map[string]any
		f          func(*smugmug.Image, error)
	}{
		{
			name:       "missing image",
			imageKey:   "VPB9RVH-0",
			expansions: []string{},
			f: func(image *smugmug.Image, err error) {
				a.Error(err)
				a.Nil(image)
				var fault *smugmug.Fault
				ok := errors.As(err, &fault)
				a.True(ok)
				a.Equal(http.StatusNotFound, fault.Code)
				a.Equal(http.StatusText(http.StatusNotFound), fault.Message)
			},
		},
		{
			name:       "single image",
			imageKey:   "VPB9RVH-0",
			filename:   "testdata/image_VPB9RVH-0.json",
			expansions: []string{},
			f: func(image *smugmug.Image, err error) {
				a.NotNil(image)
				a.NoError(err)
				a.Nil(image.Album)
			},
		},
		{
			name:       "patch image",
			imageKey:   "VPB9RVH-0",
			filename:   "testdata/image_B2fHSt7-0.json",
			expansions: []string{},
			patch:      map[string]any{"Keywords": []string{}},
			f: func(image *smugmug.Image, err error) {
				a.NotNil(image)
				a.NoError(err)
				a.NotNil(image.Album)
			},
		},
		{
			name:       "image size details expansion",
			imageKey:   "mQRcX2V-0",
			filename:   "testdata/image_mQRcX2V-0_expansions.json",
			expansions: []string{"ImageSizeDetails"},
			f: func(image *smugmug.Image, err error) {
				a.NotNil(image)
				a.NoError(err)
				a.NotNil(image.ImageSizeDetails)
			},
		},
		{
			name:       "image size details & metadata expansion",
			imageKey:   "UQpV019-0",
			filename:   "testdata/image_UQpV019_metadata.json",
			expansions: []string{"ImageMetadata"},
			f: func(image *smugmug.Image, err error) {
				a.NotNil(image)
				a.NoError(err)
				a.NotNil(image.ImageMetadata)
				a.Equal("iPhone X back dual camera 4mm f/1.8", image.ImageMetadata.Lens)
			},
		},
		{
			name:    "api option failure",
			options: []smugmug.APIOption{withError()},
			f: func(image *smugmug.Image, err error) {
				a.Nil(image)
				a.Error(err)
				a.ErrorIs(err, errFail)
			},
		},
		{
			name:     "fail with bad json",
			filename: "image_test.go",
			f: func(image *smugmug.Image, err error) {
				a.Nil(image)
				a.Error(err)
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.filename == "" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				http.ServeFile(w, r, tt.filename)
			}))
			defer svr.Close()

			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
			a.NoError(err)
			opts := []smugmug.APIOption{smugmug.WithExpansions(tt.expansions...)}
			if len(tt.options) > 0 {
				opts = append(opts, tt.options...)
			}

			var image *smugmug.Image
			if tt.patch != nil {
				image, err = mg.Image.Patch(context.TODO(), tt.imageKey, tt.patch, opts...)
				tt.f(image, err)
			} else {
				image, err = mg.Image.Image(context.TODO(), tt.imageKey, opts...)
				tt.f(image, err)
			}
		})
	}
}

func TestImages(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name       string
		albumKey   string
		expansions []string
		filename   string
		options    []smugmug.APIOption
		f          func(images []*smugmug.Image, pages *smugmug.Pages, err error)
	}{
		{
			name:     "success",
			albumKey: "WpK3n2",
			filename: "testdata/images_WpK3n2_expansions.json",
			options: []smugmug.APIOption{
				smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("Album"),
			},
			f: func(images []*smugmug.Image, pages *smugmug.Pages, err error) {
				a.NoError(err)
				a.NotNil(images)
				a.NotNil(pages)
				a.Equal(4, pages.Total)
				a.Equal("WpK3n2", images[0].Album.AlbumKey)
			},
		},
		{
			name:     "fail with api option error",
			albumKey: "WpK3n2",
			filename: "testdata/images_WpK3n2_expansions.json",
			options:  []smugmug.APIOption{withError()},
			f: func(images []*smugmug.Image, pages *smugmug.Pages, err error) {
				a.Error(err)
				a.ErrorIs(err, errFail)
				a.Nil(images)
				a.Nil(pages)
			},
		},
		{
			name:     "fail with bad json",
			albumKey: "WpK3n2",
			filename: "image_test.go",
			options:  []smugmug.APIOption{withError()},
			f: func(images []*smugmug.Image, pages *smugmug.Pages, err error) {
				a.Error(err)
				a.ErrorIs(err, errFail)
				a.Nil(images)
				a.Nil(pages)
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, tt.filename)
			}))
			defer svr.Close()

			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
			a.NoError(err)
			images, pages, err := mg.Image.Images(context.TODO(), tt.albumKey, tt.options...)
			tt.f(images, pages, err)
		})
	}
}

func TestImagesIter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/HZMsPf!images", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var i int
		ts, ok := r.URL.Query()["start"]
		a.True(ok, "did not find `start` parameter")
		a.Len(ts, 1, "expected to find a single value for `start`")
		switch ts[0] {
		case "1":
			i = 1
		case "31":
			i = 2
		default:
			a.Failf("unexpected starting value {%s}", ts[0])
		}
		http.ServeFile(w, r, fmt.Sprintf("testdata/album_images_HZMsPf_page_%d.json", i))
	}))

	svr := httptest.NewServer(mux)
	defer svr.Close()

	var n int
	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	a.NotNil(mg)
	err = mg.Image.ImagesIter(context.TODO(), "HZMsPf", func(_ *smugmug.Image) (bool, error) {
		n++
		return true, nil
	}, smugmug.WithSearch("", "Marmot"))
	a.NoError(err)
	a.Equal(34, n)
}

func TestDeleteImage(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name     string
		albumKey string
		imageKey string
		filename string
		options  []smugmug.APIOption
		f        func(bool, error)
	}{
		{
			name:     "success",
			imageKey: "VPB9RVH-0",
			filename: "testdata/image_743XwH7_delete.json",
			f: func(ok bool, err error) {
				a.NoError(err)
				a.True(ok)
			},
		},
		{
			name:     "success",
			imageKey: "VPB9RVH-0",
			filename: "testdata/image_743XwH7_delete.json",
			options:  []smugmug.APIOption{withError()},
			f: func(ok bool, err error) {
				a.Error(err)
				a.False(ok)
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.filename == "" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				http.ServeFile(w, r, tt.filename)
			}))
			defer svr.Close()

			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
			a.NoError(err)

			ctx := context.TODO()
			ok, err := mg.Image.Delete(ctx, tt.albumKey, tt.imageKey, tt.options...)
			tt.f(ok, err)
		})
	}
}
