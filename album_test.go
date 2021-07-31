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

func TestAlbum(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/album_RM4BL2.json")
		a.NoError(err)
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	user, err := mg.Album.Album(context.Background(), "RM4BL2")
	a.NoError(err)
	a.NotNil(user)
}

func TestAlbumSearch(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/album_search_marmot_page_1.json")
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	albums, pages, err := mg.Album.Search(context.Background())
	a.NoError(err)
	a.NotNil(albums)
	a.NotNil(pages)
	a.Equal(10, pages.Count)
	a.Equal(20, pages.Total)
}

func TestAlbumSearchIter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []func(*smugmug.Client) error{
		// search iteration for search results
		func(mg *smugmug.Client) error {
			var n int
			err := mg.Album.SearchIter(context.Background(), func(album *smugmug.Album) (bool, error) {
				n++
				if n == 11 {
					a.Equal("HNxNF4", album.AlbumKey)
				}
				return true, nil
			}, smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("HighlightImage"))
			a.Equal(20, n)
			return err
		},
		// album iteration for a user
		func(mg *smugmug.Client) error {
			var n int
			err := mg.Album.AlbumsIter(context.Background(), "foobar", func(album *smugmug.Album) (bool, error) {
				n++
				if n == 11 {
					a.Equal("HNxNF4", album.AlbumKey)
				}
				return true, nil
			}, smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("HighlightImage"))
			a.Equal(20, n)
			return err
		},
	}

	for i := range tests {
		f := tests[i]

		var j int
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var fn string
			switch j {
			case 0:
				fn = "testdata/album_search_marmot_page_1.json"
			case 1:
				fn = "testdata/album_search_marmot_page_2.json"
			default:
				a.Fail("expected i <= 1, not {%d}", j)
				return
			}
			fp, err := os.Open(fn)
			a.NoError(err)
			defer fp.Close()
			_, err = io.Copy(w, fp)
			a.NoError(err)
			j++
		}))
		defer svr.Close()

		mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
		a.NoError(err)
		a.NoError(f(mg))
	}
}
