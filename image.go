package smugmug

import (
	"context"
	"fmt"
	"net/http"
)

// ImageService is the API for image endpoints
type ImageService service

func (s *ImageService) Images(ctx context.Context, albumID string) ([]*Image, error) {
	uri := fmt.Sprintf("album/%s!images", albumID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri)
	if err != nil {
		return nil, err
	}
	res := &ImagesResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return res.Response.Images, err
}
