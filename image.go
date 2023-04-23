package smugmug

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// @todo(bzimmer) add image search

var errNoImage = errors.New("no image")

// ImageService is the API for image endpoints
type ImageService service

// ImageIterFunc is called for iteration of images results
type ImageIterFunc func(*Image) (bool, error)

func (s *ImageService) image(req *http.Request) (*Image, error) {
	res := &imageResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return s.expand(res.Response.Image, res.Expansions)
}

func (s *ImageService) images(req *http.Request) ([]*Image, *Pages, error) {
	res := &imagesResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	for i := range res.Response.Images {
		if _, err = s.expand(res.Response.Images[i], res.Expansions); err != nil {
			if !errors.Is(err, errNoImage) {
				return nil, nil, err
			}
		}
	}
	return res.Response.Images, res.Response.Pages, nil
}

func (s *ImageService) expand(image *Image, expansions map[string]*json.RawMessage) (*Image, error) {
	// a delete request will result in no image in the response
	if image == nil {
		return nil, errNoImage
	}
	if image.URIs.ImageSizeDetails != nil {
		if val, ok := expansions[image.URIs.ImageSizeDetails.URI]; ok {
			res := struct {
				ImageSizeDetails *ImageSizeDetails `json:"ImageSizeDetails"`
			}{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			image.ImageSizeDetails = res.ImageSizeDetails
		}
	}
	// Album exists when expanding an image by the album key (eg HighlightImage)
	if image.URIs.Album != nil {
		if val, ok := expansions[image.URIs.Album.URI]; ok {
			res := struct {
				Album *Album `json:"Album"`
			}{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			image.Album = res.Album
		}
		return image, nil
	}
	// ImageAlbum exists when querying an image directly
	if image.URIs.ImageAlbum != nil {
		if val, ok := expansions[image.URIs.ImageAlbum.URI]; ok {
			res := struct {
				Album *Album `json:"Album"`
			}{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			image.Album = res.Album
		}
		return image, nil
	}
	return image, nil
}

// Image returns the image for `imageKey`
func (s *ImageService) Image(ctx context.Context, imageKey string, options ...APIOption) (*Image, error) {
	uri := fmt.Sprintf("image/%s", imageKey)
	req, err := s.client.newRequest(ctx, uri, options)
	if err != nil {
		return nil, err
	}
	return s.image(req)
}

// Patch updates the metadata for `imageKey`
// The image is not updated; to update the image use the `Upload` service
func (s *ImageService) Patch(
	ctx context.Context, imageKey string, data map[string]interface{}, options ...APIOption) (*Image, error) {
	uri := fmt.Sprintf("image/%s", imageKey)
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := s.client.newRequestWithBody(ctx, http.MethodPatch, uri, bytes.NewReader(body), options)
	if err != nil {
		return nil, err
	}
	return s.image(req)
}

// Delete deletes the image for `imageKey` from `albumKey`
func (s *ImageService) Delete(ctx context.Context, albumKey, imageKey string, options ...APIOption) (bool, error) {
	uri := fmt.Sprintf("album/%s/image/%s", albumKey, imageKey)
	req, err := s.client.newRequestWithBody(ctx, http.MethodDelete, uri, http.NoBody, options)
	if err != nil {
		return false, err
	}
	_, err = s.image(req)
	if err == nil || errors.Is(err, errNoImage) {
		return true, nil
	}
	return false, err
}

// Images returns a single page of image results for the album
func (s *ImageService) Images(
	ctx context.Context, albumKey string, options ...APIOption) ([]*Image, *Pages, error) {
	uri := fmt.Sprintf("album/%s!images", albumKey)
	req, err := s.client.newRequest(ctx, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.images(req)
}

// ImagesIter iterates all images in the album
func (s *ImageService) ImagesIter(
	ctx context.Context, albumKey string, iter ImageIterFunc, options ...APIOption) error {
	return iterate(ctx, func(ctx context.Context, options ...APIOption) ([]*Image, *Pages, error) {
		return s.Images(ctx, albumKey, options...)
	}, iter, options...)
}

type imagesResponse struct {
	Response struct {
		Images []*Image `json:"AlbumImage"`
		Pages  *Pages   `json:"Pages"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type imageResponse struct {
	Response struct {
		Image *Image `json:"Image"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}
