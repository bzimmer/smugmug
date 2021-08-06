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
type NodeIterFunc func(*Node) (bool, error)

type nodesQueryFunc func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error)

func (s *NodeService) iter(ctx context.Context, q nodesQueryFunc, f NodeIterFunc, options ...APIOption) error {
	n := 0
	page := WithPagination(1, batchSize)
	for {
		nodes, pages, err := q(ctx, append(options, page)...)
		if err != nil {
			return err
		}
		n += pages.Count
		for _, node := range nodes {
			if ok, err := f(node); err != nil {
				return err
			} else if !ok {
				return nil
			}
		}
		if n == pages.Total {
			return nil
		}
		page = WithPagination(pages.Start+pages.Count, batchSize)
	}
}

func (s *NodeService) node(req *http.Request) (*Node, error) {
	res := &nodeResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return s.expand(res.Response.Node, res.Expansions)
}

func (s *NodeService) nodes(req *http.Request) ([]*Node, *Pages, error) {
	res := &nodesResponse{}
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
	if node.URIs.Parent != nil {
		if val, ok := expansions[node.URIs.Parent.URI]; ok {
			res := struct{ Node *Node }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			node.Parent = res.Node
		}
	}
	switch node.Type {
	case "Folder":
		// deprecated
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
	return s.node(req)
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

// ChildrenIter iterates all direct children of the node
func (s *NodeService) ChildrenIter(ctx context.Context, nodeID string, iter NodeIterFunc, options ...APIOption) error {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
		return s.Children(ctx, nodeID, options...)
	}, iter, options...)
}

// Search returns a single page of search results (does not traverse)
func (s *NodeService) Search(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "node!search", options)
	if err != nil {
		return nil, nil, err
	}
	return s.nodes(req)
}

// SearchIter iterates all search results
func (s *NodeService) SearchIter(ctx context.Context, iter NodeIterFunc, options ...APIOption) error {
	return s.iter(ctx, s.Search, iter, options...)
}

// Parent returns the parent node
func (s *NodeService) Parent(ctx context.Context, nodeID string, options ...APIOption) (*Node, error) {
	uri := fmt.Sprintf("node/%s!parent", nodeID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, err
	}
	return s.node(req)
}

// Parents returns a single page of parent nodes (does not traverse)
func (s *NodeService) Parents(ctx context.Context, nodeID string, options ...APIOption) ([]*Node, *Pages, error) {
	uri := fmt.Sprintf("node/%s!parents", nodeID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.nodes(req)
}

// ParentsIter iterates all parental ancestors
func (s *NodeService) ParentsIter(ctx context.Context, nodeID string, iter NodeIterFunc, options ...APIOption) error {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
		return s.Parents(ctx, nodeID, options...)
	}, iter, options...)
}

// Walk traverses all children of the node rooted at `nodeID`
func (s *NodeService) Walk(ctx context.Context, nodeID string, fn NodeIterFunc, options ...APIOption) error {
	k := &stack{}
	k.push(nodeID, nil)
	for {
		nid, ok := k.pop()
		if !ok {
			return nil
		}
		node := nid.node
		if node == nil {
			var err error
			node, err = s.Node(ctx, nid.id, options...)
			if err != nil {
				return err
			}
		}
		if ok, err := fn(node); err != nil {
			return err
		} else if !ok {
			return nil
		}
		switch node.Type {
		case "Album", "System Album":
			// ignore, no children
		case "Folder":
			if err := s.ChildrenIter(ctx, nid.id, func(node *Node) (bool, error) {
				k.push(node.NodeID, node)
				return true, nil
			}, options...); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unhandled type {%s}", node.Type)
		}
	}
}

type item struct {
	id   string
	node *Node
}

type stack []*item

func (s *stack) push(id string, node *Node) {
	*s = append(*s, &item{id: id, node: node})
}

func (s *stack) pop() (*item, bool) {
	if len(*s) == 0 {
		return nil, false
	}
	index := len(*s) - 1
	element := (*s)[index]
	*s = (*s)[:index]
	return element, true
}

type nodesResponse struct {
	Response struct {
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		Node           []*Node `json:"Node"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		Pages          *Pages  `json:"Pages"`
		Timing         *timing `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type nodeResponse struct {
	Response struct {
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		Node           *Node   `json:"Node"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		Timing         *timing `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}
