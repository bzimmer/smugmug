package smugmug

import (
	"context"
	"fmt"
	"net/http"
)

// AlbumService is the API for album endpoints
type AlbumService service

func (s *AlbumService) Albums(ctx context.Context, userID string) ([]*Album, error) {
	uri := fmt.Sprintf("user/%s!albums", userID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri)
	if err != nil {
		return nil, err
	}
	res := &AlbumsResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return res.Response.Album, err
}
