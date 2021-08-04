package smugmug_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
)

func TestImage(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		imageKey   string
		expansions []string
		filename   string
		f          func(*smugmug.Image, error)
	}{
		{
			imageKey:   "VPB9RVH-0",
			expansions: []string{},
			f: func(image *smugmug.Image, err error) {
				a.Error(err)
				a.Nil(image)
			},
		},
		{
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
			imageKey:   "mQRcX2V-0",
			filename:   "testdata/image_mQRcX2V-0_expansions.json",
			expansions: []string{"ImageSizeDetails"},
			f: func(image *smugmug.Image, err error) {
				a.NotNil(image)
				a.NoError(err)
				a.NotNil(image.ImageSizeDetails)
			},
		},
	}
	for i := range tests {
		test := tests[i]
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if test.filename == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			fp, err := os.Open(test.filename)
			a.NoError(err)
			defer fp.Close()
			_, err = io.Copy(w, fp)
			a.NoError(err)
		}))
		defer svr.Close()

		mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
		a.NoError(err)
		image, err := mg.Image.Image(context.Background(), test.imageKey, smugmug.WithExpansions(test.expansions...))
		test.f(image, err)
	}
}

func TestImages(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/images_WpK3n2_expansions.json")
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	images, pages, err := mg.Image.Images(context.Background(), "WpK3n2",
		smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("Album"))
	a.NoError(err)
	a.NotNil(images)
	a.NotNil(pages)
	a.Equal(4, pages.Total)
	a.Equal("WpK3n2", images[0].Album.AlbumKey)
}

func TestImagesIter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	var i int
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var fn string
		switch i {
		case 0:
			fn = "testdata/album_images_HZMsPf_page_1.json"
		case 1:
			fn = "testdata/album_images_HZMsPf_page_2.json"
		default:
			a.Fail("expected i <= 1, not {%d}", i)
			return
		}
		fp, err := os.Open(fn)
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
		i++
	}))
	defer svr.Close()

	var n int
	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	err = mg.Image.ImagesIter(context.Background(), "HZMsPf", func(img *smugmug.Image) (bool, error) {
		n++
		return true, nil
	}, smugmug.WithSearch("", "Marmot"))
	a.NoError(err)
	a.Equal(34, n)
}
