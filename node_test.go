package smugmug_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
)

func TestNode(t *testing.T) { //nolint
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name     string
		f        func(*smugmug.Node, error)
		filename string
		nodeID   string
		options  []smugmug.APIOption
		parent   bool
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
			name:     "api option failure",
			nodeID:   "zx4Fx",
			filename: "testdata/node_zx4Fx.json",
			options:  []smugmug.APIOption{withError()},
			f: func(node *smugmug.Node, err error) {
				a.Error(err)
				a.True(errors.Is(err, errFail))
				a.Nil(node)
			},
		},
		{
			nodeID:   "zx4Fx",
			filename: "testdata/node_zx4Fx.json",
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
			},
		},
		{
			nodeID:   "kTR76",
			filename: "testdata/node_kTR76.json",
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
				a.Equal("Folder", node.Type)
				a.Equal("zx4Fx", node.Parent.NodeID)
			},
		},
		{
			nodeID:   "JDVkPQ",
			filename: "testdata/node_JDVkPQ.json",
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
				a.Equal("Album", node.Type)
			},
		},
		{
			nodeID:   "ZFJQ9",
			filename: "testdata/node_ZFJQ9_parent.json",
			parent:   true,
			f: func(node *smugmug.Node, err error) {
				a.NoError(err)
				a.NotNil(node)
				a.Equal("Folder", node.Type)
				a.Equal("zx4Fx", node.NodeID)
			},
		},
		{
			nodeID:   "ZFJQ9",
			filename: "testdata/node_ZFJQ9_parent.json",
			parent:   true,
			options:  []smugmug.APIOption{withError()},
			f: func(node *smugmug.Node, err error) {
				a.Error(err)
				a.True(errors.Is(err, errFail))
				a.Nil(node)
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		if tt.name == "" {
			tt.name = tt.filename
		}
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.filename == "" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				http.ServeFile(w, r, tt.filename)
			}))
			defer svr.Close()
			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
			a.NoError(err)

			q := mg.Node.Node
			if tt.parent {
				q = mg.Node.Parent
			}
			tt.f(q(context.TODO(), tt.nodeID, tt.options...))
		})
	}
}

func TestNodes(t *testing.T) { //nolint
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		status int
		name   string
		res    map[int]string
		f      func(*smugmug.Client)
	}{
		{
			name: "search iteration for search results",
			f: func(mg *smugmug.Client) {
				var n int
				err := mg.Node.SearchIter(context.TODO(), func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, smugmug.WithSearch("", "Marmot"), smugmug.WithExpansions("HighlightImage"))
				a.Equal(19, n)
				a.NoError(err)
			},
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		{
			name: "search iteration for search results fail",
			f: func(mg *smugmug.Client) {
				var n int
				err := mg.Node.SearchIter(context.TODO(), func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, withError())
				a.Error(err)
				a.True(errors.Is(err, errFail))
			},
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		{
			name: "node iteration of children",
			f: func(mg *smugmug.Client) {
				var n int
				err := mg.Node.ChildrenIter(context.TODO(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, smugmug.WithExpansions("HighlightImage"))
				a.Equal(19, n)
				a.NoError(err)
			},
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		{
			name: "node iteration of children fail",
			f: func(mg *smugmug.Client) {
				var n int
				err := mg.Node.ChildrenIter(context.TODO(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, withError())
				a.Error(err)
				a.True(errors.Is(err, errFail))
			},
			res: map[int]string{
				0: "testdata/node_children_zx4Fx_page_1.json",
				1: "testdata/node_children_zx4Fx_page_2.json",
			},
		},
		{
			name: "node walk iteration",
			f: func(mg *smugmug.Client) {
				var n int
				err := mg.Node.Walk(context.TODO(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					return true, nil
				}, smugmug.WithExpansions("HighlightImage"))
				// one node, eleven albums
				a.Equal(11+1, n)
				a.NoError(err)
			},
			res: map[int]string{
				0: "testdata/node_zx4Fx.json",
				1: "testdata/node_children_zx4Fx_albums_page_1.json",
				2: "testdata/node_children_zx4Fx_albums_page_2.json",
			},
		},
		{
			name: "node walk iteration to defined depth",
			f: func(mg *smugmug.Client) {
				var n int
				err := mg.Node.WalkN(context.TODO(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					n++
					a.Equal("zx4Fx", node.NodeID)
					return true, nil
				}, 0)
				a.Equal(1, n)
				a.NoError(err)
			},
			res: map[int]string{
				0: "testdata/node_zx4Fx.json",
				1: "testdata/node_children_zx4Fx_albums_page_1.json",
				2: "testdata/node_children_zx4Fx_albums_page_2.json",
			},
		},
		{
			name: "node walk with type `unknown`",
			f: func(mg *smugmug.Client) {
				err := mg.Node.Walk(context.TODO(), "zx4Fx", func(node *smugmug.Node) (bool, error) {
					return true, nil
				})
				a.Error(err)
			},
			res: map[int]string{
				0: "testdata/node_zx4Fx_type_unknown.json",
			},
		},
		{
			name: "parents",
			f: func(mg *smugmug.Client) {
				var parents []string
				err := mg.Node.ParentsIter(context.TODO(), "g8CLb2", func(node *smugmug.Node) (bool, error) {
					parents = append(parents, node.NodeID)
					return true, nil
				})
				a.NoError(err)
				a.Equal(3, len(parents))
				a.Equal([]string{"g8CLb2", "T8q7k", "zx4Fx"}, parents)
			},
			res: map[int]string{
				0: "testdata/node_g8CLb2_parents.json",
			},
		},
		{
			name:   "parents fail",
			status: http.StatusForbidden,
			f: func(mg *smugmug.Client) {
				parents, _, err := mg.Node.Parents(context.TODO(), "g8CLb2")
				a.Error(err)
				a.Nil(parents)
			},
			res: map[int]string{
				0: "testdata/node_g8CLb2_parents.json",
			},
		},
		{
			name: "parents fail with api option",
			f: func(mg *smugmug.Client) {
				var parents []string
				err := mg.Node.ParentsIter(context.TODO(), "g8CLb2", func(node *smugmug.Node) (bool, error) {
					parents = append(parents, node.NodeID)
					return true, nil
				}, withError())
				a.Error(err)
				a.True(errors.Is(err, errFail))
			},
			res: map[int]string{
				0: "testdata/node_g8CLb2_parents.json",
			},
		},
		{
			name: "node creation",
			f: func(mg *smugmug.Client) {
				nodelet := &smugmug.Nodelet{
					Name:    "foobar",
					URLName: "Foobar",
				}
				node, err := mg.Node.Create(context.TODO(), "g8CLb2", nodelet)
				a.NoError(err)
				a.NotNil(node)
				a.Equal("xmQnCV", node.NodeID)
			},
			res: map[int]string{
				0: "testdata/node_create_xmQnCV_.json",
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
				http.ServeFile(w, r, fn)
				j++
			}))
			defer svr.Close()

			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
			a.NoError(err)
			test.f(mg)
		})
	}
}
