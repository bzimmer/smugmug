package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// NodeService is the API for node endpoints
type NodeService service

type nodesCallback func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error)

func (s *NodeService) iter(ctx context.Context, f nodesCallback, options ...APIOption) ([]*Node, error) {
	var res []*Node
	page := WithPagination(1, batchSize)
	for {
		nodes, pages, err := f(ctx, append(options, page)...)
		if err != nil {
			return nil, err
		}
		res = append(res, nodes...)
		if len(res) == pages.Total {
			return res, nil
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
	for key, val := range expansions {
		switch {
		case key == node.URIs.User.URI:
			res := struct{ User *User }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			node.User = res.User
		case key == node.URIs.HighlightImage.URI:
			res := struct{ Image *Image }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			node.HighlightImage = res.Image
		case key == node.URIs.ParentNode.URI:
			res := struct{ Node *Node }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			node.Parent = res.Node
		default:
			switch node.Type {
			case "Folder":
				if key == node.URIs.FolderByID.URI {
					res := struct{ Folder *Folder }{}
					if err := json.Unmarshal(*val, &res); err != nil {
						return nil, err
					}
					node.Folder = res.Folder
				}
			case "Album":
				if key == node.URIs.Album.URI {
					res := struct{ Album *Album }{}
					if err := json.Unmarshal(*val, &res); err != nil {
						return nil, err
					}
					node.Album = res.Album
				}
			}
		}
	}
	return node, nil
}

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

func (s *NodeService) Children(ctx context.Context, nodeID string, options ...APIOption) ([]*Node, *Pages, error) {
	uri := fmt.Sprintf("node/%s!children", nodeID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.nodes(req)
}

func (s *NodeService) ChildrenAll(ctx context.Context, nodeID string, options ...APIOption) ([]*Node, error) {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
		return s.Children(ctx, nodeID, options...)
	}, options...)
}

func (s *NodeService) Search(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
	uri := "node!search"
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.nodes(req)
}

func (s *NodeService) SearchAll(ctx context.Context, nodeID string, options ...APIOption) ([]*Node, error) {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
		return s.Search(ctx, options...)
	}, options...)
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

type WalkFunc func(context.Context, *Node) error

func (s *NodeService) Walk(ctx context.Context, nodeID string, fn WalkFunc, options ...APIOption) error {
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
		if err := fn(ctx, node); err != nil {
			return err
		}
		switch node.Type {
		case "Album":
		case "Folder":
			fallthrough
		default:
			children, err := s.ChildrenAll(ctx, nid, options...)
			if err != nil {
				return err
			}
			for _, child := range children {
				k.Push(child.NodeID)
			}
		}
	}
}
