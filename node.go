package smugmug

import (
	"context"
	"net/http"
)

// NodeService is the API for node endpoints
type NodeService service

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
