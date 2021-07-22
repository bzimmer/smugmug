package smugmug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AlbumService is the API for album endpoints
type AlbumService service

func (s *AlbumService) expand(album *Album, expansions map[string]*json.RawMessage) (*Album, error) {
	for key, val := range expansions {
		switch {
		case key == album.URIs.User.URI:
			res := struct{ User *User }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			album.User = res.User
		case key == album.URIs.AlbumHighlightImage.URI: // deprecated but supported
			res := struct{ AlbumImage *Image }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			album.HighlightImage = res.AlbumImage
		case key == album.URIs.HighlightImage.URI:
			res := struct{ Image *Image }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			album.HighlightImage = res.Image
		case key == album.URIs.Node.URI:
			res := struct{ Node *Node }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			album.Node = res.Node
		case album.URIs.Folder != nil && key == album.URIs.Folder.URI:
			res := struct{ Folder *Folder }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			album.Folder = res.Folder
		default:
		}
	}
	return album, nil
}

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

type albumsCallback func(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error)

func (s *AlbumService) iter(ctx context.Context, f albumsCallback, options ...APIOption) ([]*Album, error) {
	var res []*Album
	page := WithPagination(1, batchSize)
	for {
		albums, pages, err := f(ctx, append(options, page)...)
		if err != nil {
			return nil, err
		}
		res = append(res, albums...)
		if len(res) == pages.Total {
			return res, nil
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

func (s *AlbumService) Albums(ctx context.Context, userID string, options ...APIOption) ([]*Album, *Pages, error) {
	uri := fmt.Sprintf("user/%s!albums", userID)
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.albums(req)
}

func (s *AlbumService) AlbumsAll(ctx context.Context, userID string, options ...APIOption) ([]*Album, error) {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
		return s.Albums(ctx, userID, options...)
	}, options...)
}

func (s *AlbumService) Search(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
	uri := "album!search"
	req, err := s.client.newRequest(ctx, http.MethodGet, uri, options)
	if err != nil {
		return nil, nil, err
	}
	return s.albums(req)
}

func (s *AlbumService) SearchAll(ctx context.Context, options ...APIOption) ([]*Album, error) {
	return s.iter(ctx, func(ctx context.Context, options ...APIOption) ([]*Album, *Pages, error) {
		return s.Search(ctx, options...)
	}, options...)
}
