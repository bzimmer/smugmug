package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ImageService is the API for image endpoints
type ImageService service

func (s *ImageService) Image(ctx context.Context, imageKey string, options ...APIOption) (*Image, error) {
	uri := fmt.Sprintf("image/%s", imageKey)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, err
	}
	res := &ImageResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}

	if res.Expansions != nil {
		res.Response.Image.Expansions = &ImageExpansions{}
		for key, val := range res.Expansions {
			switch {
			case strings.HasSuffix(key, "!sizedetails"):
				type exp struct {
					ImageSizeDetails *ImageSizeDetails `json:"ImageSizeDetails"`
				}
				t := &exp{}
				err := json.Unmarshal(*val, t)
				if err != nil {
					return nil, err
				}
				res.Response.Image.Expansions.ImageSizeDetails = t.ImageSizeDetails
			}
		}
	}

	return res.Response.Image, err
}

func (s *ImageService) Images(ctx context.Context, albumID string, options ...APIOption) ([]*Image, *Pages, error) {
	uri := fmt.Sprintf("album/%s!images", albumID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}

	res := &ImagesResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}

	return res.Response.Images, res.Response.Pages, err
}
