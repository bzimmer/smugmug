package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AlbumService is the API for album endpoints
type AlbumService service

func (s *AlbumService) Album(ctx context.Context, albumKey string, options ...APIOption) (*Album, error) {
	uri := fmt.Sprintf("album/%s", albumKey)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, err
	}
	res := &AlbumResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}

	if res.Expansions != nil {
		res.Response.Album.Expansions = &AlbumExpansions{}
		for key, val := range res.Expansions {
			switch {
			case strings.HasSuffix(key, "!highlightimage"):
				type exp struct {
					Image *Image `json:"AlbumImage"`
				}
				t := &exp{}
				err := json.Unmarshal(*val, t)
				if err != nil {
					return nil, err
				}
				res.Response.Album.Expansions.HighlightImage = t.Image
			case strings.HasSuffix(key, "!images"):
				type exp struct {
					Images []*Image `json:"AlbumImage"`
				}
				t := &exp{}
				err := json.Unmarshal(*val, t)
				if err != nil {
					return nil, err
				}
				res.Response.Album.Expansions.Images = t.Images
			}
		}
	}

	return res.Response.Album, err
}

func (s *AlbumService) Albums(ctx context.Context, userID string, options ...APIOption) ([]*Album, *Pages, error) {
	uri := fmt.Sprintf("user/%s!albums", userID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	res := &AlbumsResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	return res.Response.Album, res.Response.Pages, nil
}

func (s *AlbumService) Search(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
	uri := "album!search"
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	res := &AlbumsResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	return res.Response.Album, res.Response.Pages, nil
}
