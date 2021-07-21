package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NodeService is the API for node endpoints
type NodeService service

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

	if res.Expansions != nil {
		res.Response.Node.Expansions = &NodeExpansions{}
		for key, val := range res.Expansions {
			switch {
			case strings.HasSuffix(key, "!children"):
				type exp struct {
					Node []*Node `json:"Node"`
				}
				t := &exp{}
				err := json.Unmarshal(*val, t)
				if err != nil {
					return nil, err
				}
				res.Response.Node.Expansions.ChildNodes = t.Node
			case strings.HasPrefix(key, "/api/v2/album/"):
				type exp struct {
					Album *Album `json:"Album"`
				}
				t := &exp{}
				err := json.Unmarshal(*val, t)
				if err != nil {
					return nil, err
				}
				res.Response.Node.Expansions.Album = t.Album
			}
		}
	}

	return res.Response.Node, err
}

func (s *NodeService) Search(ctx context.Context, options ...APIOption) ([]*Node, *Pages, error) {
	uri := "node!search"
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	res := &NodesResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	return res.Response.Node, res.Response.Pages, nil
}
