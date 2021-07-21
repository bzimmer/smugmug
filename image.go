package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ImageService is the API for image endpoints
type ImageService service

type imagesCallback func(ctx context.Context, options ...APIOption) ([]*Image, *Pages, error)

func (s *ImageService) iter(ctx context.Context, f imagesCallback, options ...APIOption) ([]*Image, error) {
	var res []*Image
	page := WithPagination(1, batchSize)
	for {
		images, pages, err := f(ctx, append(options, page)...)
		if err != nil {
			return nil, err
		}
		res = append(res, images...)
		if len(res) == pages.Total {
			return res, nil
		}
		page = WithPagination(pages.Start+pages.Count, batchSize)
	}
}

func (s *ImageService) images(req *http.Request) ([]*Image, *Pages, error) {
	res := &ImagesResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	for i := range res.Response.Images {
		if _, err := s.expand(res.Response.Images[i], res.Expansions); err != nil {
			return nil, nil, err
		}
	}
	return res.Response.Images, res.Response.Pages, err
}

func (s *ImageService) expand(image *Image, expansions map[string]*json.RawMessage) (*Image, error) {
	for key, val := range expansions {
		switch {
		case key == image.URIs.ImageSizeDetails.URI:
			res := struct{ ImageSizeDetails *ImageSizeDetails }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			image.ImageSizeDetails = res.ImageSizeDetails
		default:
		}
	}
	return image, nil
}

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
	return s.expand(res.Response.Image, res.Expansions)
}

func (s *ImageService) Images(ctx context.Context, albumID string, options ...APIOption) ([]*Image, *Pages, error) {
	uri := fmt.Sprintf("album/%s!images", albumID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.images(req)
}

func (s *ImageService) ImagesAll(ctx context.Context, albumID string, options ...APIOption) ([]*Image, error) {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Image, *Pages, error) {
		return s.Images(ctx, albumID, options...)
	}, options...)
}
