package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ImageService is the API for image endpoints
type ImageService service

// ImageIterFunc is called for iteration of images results
type ImageIterFunc func(*Image) error

type imagesQueryFunc func(ctx context.Context, options ...APIOption) ([]*Image, *Pages, error)

func (s *ImageService) iter(ctx context.Context, q imagesQueryFunc, f ImageIterFunc, options ...APIOption) error {
	i := 0
	page := WithPagination(1, batchSize)
	for {
		images, pages, err := q(ctx, append(options, page)...)
		if err != nil {
			return err
		}
		i += pages.Count
		for _, image := range images {
			if err := f(image); err != nil {
				return err
			}
		}
		if i == pages.Total {
			return nil
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
	if val, ok := expansions[image.URIs.ImageSizeDetails.URI]; ok {
		res := struct{ ImageSizeDetails *ImageSizeDetails }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		image.ImageSizeDetails = res.ImageSizeDetails
	}
	return image, nil
}

// Image returns the image for `imageKey`
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

// Images returns a single page of image results for the album
func (s *ImageService) Images(ctx context.Context, albumID string, options ...APIOption) ([]*Image, *Pages, error) {
	uri := fmt.Sprintf("album/%s!images", albumID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.images(req)
}

// ImagesIter iterates all images in the album
func (s *ImageService) ImagesIter(ctx context.Context, albumID string, iter ImageIterFunc, options ...APIOption) error {
	q := func(ctx context.Context, options ...APIOption) ([]*Image, *Pages, error) {
		return s.Images(ctx, albumID, options...)
	}
	return s.iter(ctx, q, iter, options...)
}
