package smugmug

import (
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

func (s *AlbumService) expand(album *Album, expansions map[string]*json.RawMessage) (*Album, error) {
	if val, ok := expansions[album.URIs.User.URI]; ok {
		res := struct{ User *User }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		album.User = res.User
	}
	if val, ok := expansions[album.URIs.AlbumHighlightImage.URI]; ok { // deprecated but supported
		res := struct{ AlbumImage *Image }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		album.HighlightImage = res.AlbumImage
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
	if album.URIs.Folder != nil {
		if val, ok := expansions[album.URIs.Folder.URI]; ok {
			res := struct{ Folder *Folder }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			album.Folder = res.Folder
		}
	}
	return album, nil
}

// Album returns the album with key `albumKey`
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
	return s.expand(res.Response.Album, res.Expansions)
}

func (s *AlbumService) iter(ctx context.Context, q albumsQueryFunc, f AlbumIterFunc, options ...APIOption) error {
	i := 0
	page := WithPagination(1, batchSize)
	for {
		albums, pages, err := q(ctx, append(options, page)...)
		if err != nil {
			return err
		}
		i += pages.Count
		for _, album := range albums {
			if ok, err := f(album); err != nil {
				return err
			} else if !ok {
				return nil
			}
		}
		if i == pages.Total {
			return nil
		}
		page = WithPagination(pages.Start+pages.Count, batchSize)
	}
}

func (s *AlbumService) albums(req *http.Request) ([]*Album, *Pages, error) {
	res := &AlbumsResponse{}
	err := s.client.do(req, res)
	if err != nil {
		return nil, nil, err
	}
	for i := range res.Response.Album {
		if _, err := s.expand(res.Response.Album[i], res.Expansions); err != nil {
			return nil, nil, err
		}
	}
	return res.Response.Album, res.Response.Pages, err
}

// Albums returns a single page of albums for the user
func (s *AlbumService) Albums(ctx context.Context, userID string, options ...APIOption) ([]*Album, *Pages, error) {
	uri := fmt.Sprintf("user/%s!albums", userID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.albums(req)
}

// Albums iterates all albums for the user
func (s *AlbumService) AlbumsIter(ctx context.Context, userID string, iter AlbumIterFunc, options ...APIOption) error {
	q := func(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
		return s.Albums(ctx, userID, options...)
	}
	return s.iter(ctx, q, iter, options...)
}

// Search returns a single page of search results
func (s *AlbumService) Search(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
	uri := "album!search"
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
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
