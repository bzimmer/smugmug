package smugmug

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AlbumService is the API for album endpoints
type AlbumService service

// AlbumIterFunc is called for each album in the results
type AlbumIterFunc func(*Album) (bool, error)

type albumsQueryFunc func(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error)

func (s *AlbumService) album(req *http.Request) (*Album, error) {
	res := &albumResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return s.expand(res.Response.Album, res.Expansions)
}

func (s *AlbumService) expand(album *Album, expansions map[string]*json.RawMessage) (*Album, error) {
	if val, ok := expansions[album.URIs.User.URI]; ok {
		res := struct{ User *User }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		album.User = res.User
	}
	if val, ok := expansions[album.URIs.HighlightImage.URI]; ok {
		res := struct{ Image *Image }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		album.HighlightImage = res.Image
	}
	if val, ok := expansions[album.URIs.Node.URI]; ok {
		res := struct{ Node *Node }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		album.Node = res.Node
	}
	return album, nil
}

// Album returns the album with key `albumKey`
func (s *AlbumService) Album(ctx context.Context, albumKey string, options ...APIOption) (*Album, error) {
	uri := fmt.Sprintf("album/%s", albumKey)
	req, err := s.client.newRequest(ctx, uri, options)
	if err != nil {
		return nil, err
	}
	return s.album(req)
}

func (s *AlbumService) iter(ctx context.Context, q albumsQueryFunc, f AlbumIterFunc, options ...APIOption) error { //nolint
	n := 0
	page := WithPagination(1, batch)
	for {
		albums, pages, err := q(ctx, append(options, page)...)
		if err != nil {
			return err
		}
		n += pages.Count
		for _, album := range albums {
			if ok, err := f(album); err != nil {
				return err
			} else if !ok {
				return nil
			}
		}
		if n == pages.Total {
			return nil
		}
		page = WithPagination(pages.Start+pages.Count, batch)
	}
}

func (s *AlbumService) albums(req *http.Request) ([]*Album, *Pages, error) {
	res := &albumsResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	for i := range res.Response.Album {
		if _, err := s.expand(res.Response.Album[i], res.Expansions); err != nil {
			return nil, nil, err
		}
	}
	return res.Response.Album, res.Response.Pages, nil
}

// Albums returns a single page of albums for the user
func (s *AlbumService) Albums(ctx context.Context, userID string, options ...APIOption) ([]*Album, *Pages, error) {
	uri := fmt.Sprintf("user/%s!albums", userID)
	req, err := s.client.newRequest(ctx, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.albums(req)
}

// AlbumsIter iterates all albums for the user
func (s *AlbumService) AlbumsIter(ctx context.Context, userID string, iter AlbumIterFunc, options ...APIOption) error {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
		return s.Albums(ctx, userID, options...)
	}, iter, options...)
}

// Search returns a single page of search results
func (s *AlbumService) Search(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
	uri := "album!search"
	req, err := s.client.newRequest(ctx, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.albums(req)
}

// SearchIter iterates all search results
// The results of this query might be very large depending on the scope and query
func (s *AlbumService) SearchIter(ctx context.Context, iter AlbumIterFunc, options ...APIOption) error {
	return s.iter(ctx, s.Search, iter, options...)
}

// Patch updates the metadata for `albumKey`
func (s *AlbumService) Patch(ctx context.Context, albumKey string, data map[string]interface{}, options ...APIOption) (*Album, error) {
	uri := fmt.Sprintf("album/%s", albumKey)
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := s.client.newRequestWithBody(ctx, http.MethodPatch, uri, bytes.NewReader(body), options)
	if err != nil {
		return nil, err
	}
	return s.album(req)
}

type albumsResponse struct {
	Response struct {
		Album []*Album `json:"Album"`
		Pages *Pages   `json:"Pages"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type albumResponse struct {
	Response struct {
		Album *Album `json:"Album"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}
