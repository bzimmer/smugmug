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

	tests := []struct {
		albumKey string
		filename string
		f        func(*smugmug.Album, error)
	}{
		{
			albumKey: "RM4BL2",
			f: func(album *smugmug.Album, err error) {
				a.Error(err)
				a.Nil(album)
			},
		},
		{
			albumKey: "RM4BL2",
			filename: "testdata/album_RM4BL2.json",
			f: func(album *smugmug.Album, err error) {
				a.NoError(err)
				a.NotNil(album)
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
			_, err = io.Copy(w, fp)
			a.NoError(err)
		}))
		defer svr.Close()

		mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
		a.NoError(err)
		test.f(mg.Album.Album(context.TODO(), test.albumKey))
	}
}

func TestAlbumExpansions(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/album_2XrGxm_expansions.json")
		a.NoError(err)
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	album, err := mg.Album.Album(context.TODO(), "2XrGxm",
		smugmug.WithExpansions("Node", "AlbumHighlightImage", "AlbumImage", "User"))
	a.NoError(err)
	a.NotNil(album)
	a.NotNil(album.Node)
	a.NotNil(album.User)
	a.Equal("cmac", album.User.NickName)
	a.NotNil(album.HighlightImage)
	a.Equal("7952669755", album.HighlightImage.UploadKey)
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
	albums, pages, err := mg.Album.Search(context.TODO())
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
			err := mg.Album.SearchIter(context.TODO(), func(album *smugmug.Album) (bool, error) {
				n++
				if n == 11 {
					a.Equal("HNxNF4", album.AlbumKey)
				}
				return true, nil
			}, smugmug.WithSearch("", "Marmot"))
			a.Equal(20, n)
			return err
		},
		// album iteration for a user
		func(mg *smugmug.Client) error {
			var n int
			err := mg.Album.AlbumsIter(context.TODO(), "foobar", func(album *smugmug.Album) (bool, error) {
				n++
				if n == 11 {
					a.Equal("HNxNF4", album.AlbumKey)
				}
				return true, nil
			}, smugmug.WithSearch("", "Marmot"))
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
