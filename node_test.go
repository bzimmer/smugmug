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

func TestNode(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/node_zx4Fx.json")
		a.NoError(err)
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	user, err := mg.Node.Node(context.Background(), "zx4Fx")
	a.NoError(err)
	a.NotNil(user)
}

func TestNodes(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		f   func(*smugmug.Client) error
		res map[int]string
	}{
		// search iteration for search results
		{
			f: func(mg *smugmug.Client) error {
				var n int
				err := mg.Node.SearchIter(context.Background(), func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("HighlightImage"))
				a.Equal(19, n)
				return err
			},
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		// node iteration of children
		{
			f: func(mg *smugmug.Client) error {
				var n int
				err := mg.Node.ChildrenIter(context.Background(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, smugmug.WithExpansions("HighlightImage"))
				a.Equal(19, n)
				return err
			},
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		// node walk iteration
		{
			f: func(mg *smugmug.Client) error {
				var n int
				err := mg.Node.Walk(context.Background(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, smugmug.WithExpansions("HighlightImage"))
				// one node, eleven albums
				a.Equal(11+1, n)
				return err
			},
			res: map[int]string{
				0: "testdata/node_zx4Fx.json",
				1: "testdata/node_children_zx4Fx_albums_page_1.json",
				2: "testdata/node_children_zx4Fx_albums_page_2.json",
			},
		},
	}

	for i := range tests {
		test := tests[i]

		var j int
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn, ok := test.res[j]
			a.True(ok, "missing file for iteration {%d}", j)
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
		a.NoError(test.f(mg))
	}
}
