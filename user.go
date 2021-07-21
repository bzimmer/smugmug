package smugmug

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserService is the API for user endpoints
type UserService service

func (s *UserService) expand(user *User, expansions map[string]*json.RawMessage) (*User, error) {
	for key, val := range expansions {
		switch {
		case key == user.URIs.Node.URI:
			res := struct{ Node *Node }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			user.Node = res.Node
		case key == user.URIs.Folder.URI:
			res := struct{ Folder *Folder }{}
			if err := json.Unmarshal(*val, &res); err != nil {
				return nil, err
			}
			user.Folder = res.Folder
		default:
		}
	}
	return user, nil
}

func (s *UserService) User(ctx context.Context, options ...APIOption) (*User, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "!authuser", options)
	if err != nil {
		return nil, err
	}
	res := &UserResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return s.expand(res.Response.User, res.Expansions)
}
