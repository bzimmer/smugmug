package smugmug_test

import (
	"context"
	"fmt"
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

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/image_VPB9RVH-0.json")
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	user, err := mg.Image.Image(context.Background(), "VPB9RVH-0")
	a.NoError(err)
	a.NotNil(user)
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
		fmt.Println(n)
		return true, nil
	}, smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("HighlightImage"))
	a.NoError(err)
	a.Equal(34, n)
}
