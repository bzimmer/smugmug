package smugmug_test

import (
	"context"
	"errors"
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

	tests := []struct {
		f       func(*smugmug.Node, error)
		fn      string
		name    string
		nodeID  string
		options []smugmug.APIOption
		parent  bool
	}{
		{
			name:   "no response",
			nodeID: "zx4Fx",
			f: func(node *smugmug.Node, err error) {
				a.Error(err)
				a.Nil(node)
			},
		},
		{
			name:    "api option failure",
			nodeID:  "zx4Fx",
			fn:      "testdata/node_zx4Fx.json",
			options: []smugmug.APIOption{withError(true)},
			f: func(node *smugmug.Node, err error) {
				a.Error(err)
				a.True(errors.Is(err, withErr))
				a.Nil(node)
			},
		},
		{
			nodeID: "zx4Fx",
			fn:     "testdata/node_zx4Fx.json",
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
			},
		},
		{
			nodeID: "kTR76",
			fn:     "testdata/node_kTR76.json",
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
				a.Equal("Folder", node.Type)
				a.Equal("zx4Fx", node.Parent.NodeID)
			},
		},
		{
			nodeID: "JDVkPQ",
			fn:     "testdata/node_JDVkPQ.json",
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
				a.Equal("Album", node.Type)
			},
		},
		{
			nodeID: "ZFJQ9",
			fn:     "testdata/node_ZFJQ9_parent.json",
			parent: true,
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
				a.Equal("Folder", node.Type)
				a.Equal("zx4Fx", node.NodeID)
			},
		},
		{
			nodeID:  "ZFJQ9",
			fn:      "testdata/node_ZFJQ9_parent.json",
			parent:  true,
			options: []smugmug.APIOption{withError(true)},
			f: func(node *smugmug.Node, err error) {
				a.Error(err)
				a.True(errors.Is(err, withErr))
				a.Nil(node)
			},
		},
	}

	for i := range tests {
		test := tests[i]
		if test.name == "" {
			test.name = test.fn
		}
		t.Run(test.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.fn == "" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				fp, err := os.Open(test.fn)
				a.NoError(err)
				_, err = io.Copy(w, fp)
				a.NoError(err)
			}))
			defer svr.Close()
			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
			a.NoError(err)

			q := mg.Node.Node
			if test.parent {
				q = mg.Node.Parent
			}
			test.f(q(context.Background(), test.nodeID, test.options...))
		})
	}
}

func TestNodes(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name   string
		status int
		fail   bool
		res    map[int]string
		f      func(*smugmug.Client) error
	}{
		{
			name: "search iteration for search results",
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
		{
			name: "search iteration for search results fail",
			f: func(mg *smugmug.Client) error {
				var n int
				err := mg.Node.SearchIter(context.Background(), func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, withError(true))
				a.Error(err)
				a.True(errors.Is(err, withErr))
				return err
			},
			fail: true,
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		{
			name: "node iteration of children",
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
		{
			name: "node iteration of children fail",
			f: func(mg *smugmug.Client) error {
				var n int
				err := mg.Node.ChildrenIter(context.Background(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, withError(true))
				a.Error(err)
				a.True(errors.Is(err, withErr))
				return err
			},
			fail: true,
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		{
			name: "node walk iteration",
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
		{
			name: "node walk with type `unknown`",
			fail: true,
			f: func(mg *smugmug.Client) error {
				return mg.Node.Walk(context.Background(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					return true, nil
				})
			},
			res: map[int]string{
				0: "testdata/node_zx4Fx_type_unknown.json",
			},
		},
		{
			name: "parents",
			f: func(mg *smugmug.Client) error {
				var parents []string
				err := mg.Node.ParentsIter(context.Background(), "g8CLb2", func(node *smugmug.Node) (bool, error) {
					parents = append(parents, node.NodeID)
					return true, nil
				})
				a.Equal(3, len(parents))
				a.Equal([]string{"g8CLb2", "T8q7k", "zx4Fx"}, parents)
				return err
			},
			res: map[int]string{
				0: "testdata/node_g8CLb2_parents.json",
			},
		},
		{
			name:   "parents fail",
			status: http.StatusForbidden,
			f: func(mg *smugmug.Client) error {
				parents, _, err := mg.Node.Parents(context.Background(), "g8CLb2")
				a.Error(err)
				a.Nil(parents)
				return nil
			},
			res: map[int]string{
				0: "testdata/node_g8CLb2_parents.json",
			},
		},
		{
			name: "parents fail with api option",
			f: func(mg *smugmug.Client) error {
				var parents []string
				err := mg.Node.ParentsIter(context.Background(), "g8CLb2", func(node *smugmug.Node) (bool, error) {
					parents = append(parents, node.NodeID)
					return true, nil
				}, withError(true))
				a.Error(err)
				a.True(errors.Is(err, withErr))
				return err
			},
			fail: true,
			res: map[int]string{
				0: "testdata/node_g8CLb2_parents.json",
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var j int
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.status != 0 {
					w.WriteHeader(test.status)
					return
				}
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
			err = test.f(mg)
			if test.fail {
				a.Error(err)
			} else {
				a.NoError(err)
			}
		})
	}
}
