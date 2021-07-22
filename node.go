package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// NodeService is the API for node endpoints
type NodeService service

// NodeIterFunc is called for each node in the results
type NodeIterFunc func(*Node) error

type nodesQueryFunc func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error)

func (s *NodeService) iter(ctx context.Context, q nodesQueryFunc, f NodeIterFunc, options ...APIOption) error {
	i := 0
	page := WithPagination(1, batchSize)
	for {
		nodes, pages, err := q(ctx, append(options, page)...)
		if err != nil {
			return err
		}
		i += pages.Count
		for _, node := range nodes {
			if err := f(node); err != nil {
				return err
			}
		}
		if i == pages.Total {
			return nil
		}
		page = WithPagination(pages.Start+pages.Count, batchSize)
	}
}

func (s *NodeService) nodes(req *http.Request) ([]*Node, *Pages, error) {
	res := &NodesResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	for i := range res.Response.Node {
		if _, err := s.expand(res.Response.Node[i], res.Expansions); err != nil {
			return nil, nil, err
		}
	}
	return res.Response.Node, res.Response.Pages, err
}

func (s *NodeService) expand(node *Node, expansions map[string]*json.RawMessage) (*Node, error) {
	if val, ok := expansions[node.URIs.User.URI]; ok {
		res := struct{ User *User }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		node.User = res.User
	}
	if val, ok := expansions[node.URIs.HighlightImage.URI]; ok {
		res := struct{ Image *Image }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		node.HighlightImage = res.Image
	}
	if val, ok := expansions[node.URIs.ParentNode.URI]; ok {
		res := struct{ Node *Node }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		node.Parent = res.Node
	}
	switch node.Type {
	case "Folder":
		if val, ok := expansions[node.URIs.FolderByID.URI]; ok {
			res := struct{ Folder *Folder }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			node.Folder = res.Folder
		}
	case "Album":
		if val, ok := expansions[node.URIs.Album.URI]; ok {
			res := struct{ Album *Album }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			node.Album = res.Album
		}
	}
	return node, nil
}

// Node returns the node with id `nodeID`
func (s *NodeService) Node(ctx context.Context, nodeID string, options ...APIOption) (*Node, error) {
	uri := fmt.Sprintf("node/%s", nodeID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, err
	}
	res := &NodeResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return s.expand(res.Response.Node, res.Expansions)
}

// Children returns a single page of direct children of the node (does not traverse)
func (s *NodeService) Children(ctx context.Context, nodeID string, options ...APIOption) ([]*Node, *Pages, error) {
	uri := fmt.Sprintf("node/%s!children", nodeID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.nodes(req)
}

// ChildrenIter iterates all direct children of the node (does not traverse)
func (s *NodeService) ChildrenIter(ctx context.Context, nodeID string, iter NodeIterFunc, options ...APIOption) error {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
		return s.Children(ctx, nodeID, options...)
	}, iter, options...)
}

// Search returns a single page of search results (does not traverse)
func (s *NodeService) Search(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
	uri := "node!search"
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.nodes(req)
}

// Search iterates all search results (does not traverse)
func (s *NodeService) SearchIter(ctx context.Context, iter NodeIterFunc, options ...APIOption) error {
	return s.iter(ctx, s.Search, iter, options...)
}

// Walk traverses all children of the node rooted at `nodeID`
func (s *NodeService) Walk(ctx context.Context, nodeID string, fn NodeIterFunc, options ...APIOption) error {
	k := &stack{}
	k.Push(nodeID)
	for {
		nid, ok := k.Pop()
		if !ok {
			return nil
		}
		node, err := s.Node(ctx, nid, options...)
		if err != nil {
			return err
		}
		if err := fn(node); err != nil {
			return err
		}
		switch node.Type {
		case "Album":
			// ignore, no children
		case "Folder":
			if err := s.ChildrenIter(ctx, nid, func(node *Node) error {
				k.Push(node.NodeID)
				return nil
			}, options...); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unhandled type {%s}", node.Type)
		}
	}
}

type stack []string

func (s *stack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *stack) Push(str string) {
	*s = append(*s, str)
}

func (s *stack) Pop() (string, bool) {
	if s.IsEmpty() {
		return "", false
	} else {
		index := len(*s) - 1
		element := (*s)[index]
		*s = (*s)[:index]
		return element, true
	}
}
