package smugmug_test

import (
	"context"
	"encoding/json"
	"errors"
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
		name     string
		albumKey string
		filename string
		patch    map[string]interface{}
		f        func(*smugmug.Album, error)
	}{
		{
			name:     "no response",
			albumKey: "RM4BL2",
			f: func(album *smugmug.Album, err error) {
				a.Error(err)
				a.Nil(album)
			},
		},
		{
			// @see(https://github.com/bzimmer/smugmug/issues/34)
			name:     "album with no user uri",
			albumKey: "WJvpCp",
			filename: "testdata/album_WJvpCp_no_user_uri.json",
			f: func(album *smugmug.Album, err error) {
				a.NoError(err)
				a.NotNil(album)
			},
		},
		{
			name:     "valid query",
			albumKey: "RM4BL2",
			filename: "testdata/album_RM4BL2.json",
			f: func(album *smugmug.Album, err error) {
				a.NoError(err)
				a.NotNil(album)
			},
		},
		{
			name:     "valid patch",
			albumKey: "RM4BL2",
			filename: "testdata/album_RM4BL2.json",
			patch: map[string]interface{}{
				"Name": "Foo",
			},
			f: func(album *smugmug.Album, err error) {
				a.NoError(err)
				a.NotNil(album)
			},
		},
	}

	for i := range tests {
		tt := tests[i]

		mux := http.NewServeMux()
		mux.HandleFunc("/album/WJvpCp", func(w http.ResponseWriter, r *http.Request) {
			fp, err := os.Open(tt.filename)
			a.NoError(err)
			_, err = io.Copy(w, fp)
			a.NoError(err)
		})
		mux.HandleFunc("/album/RM4BL2", func(w http.ResponseWriter, r *http.Request) {
			if tt.filename == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if tt.patch != nil {
				data := make(map[string]interface{})
				dec := json.NewDecoder(r.Body)
				a.NoError(dec.Decode(&data))
				a.Contains(data, "Name")
				a.Equal(data["Name"], "Foo")
			}
			fp, err := os.Open(tt.filename)
			a.NoError(err)
			_, err = io.Copy(w, fp)
			a.NoError(err)
		})
		svr := httptest.NewServer(mux)
		defer svr.Close()

		mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
		a.NoError(err)

		switch {
		case tt.patch != nil:
			tt.f(mg.Album.Patch(context.Background(), tt.albumKey, tt.patch))
		default:
			tt.f(mg.Album.Album(context.Background(), tt.albumKey))
		}
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
		// album iteration with error
		func(mg *smugmug.Client) error {
			err := mg.Album.AlbumsIter(context.TODO(), "foobar", func(album *smugmug.Album) (bool, error) {
				return false, errors.New("dummy error")
			}, smugmug.WithSearch("", "Marmot"))
			a.Error(err)
			return nil
		},
		// album iteration with no error but no continue
		func(mg *smugmug.Client) error {
			var n int
			err := mg.Album.AlbumsIter(context.TODO(), "foobar", func(album *smugmug.Album) (bool, error) {
				n++
				return false, nil
			}, smugmug.WithSearch("", "Marmot"))
			a.NoError(err)
			a.Equal(1, n)
			return nil
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
