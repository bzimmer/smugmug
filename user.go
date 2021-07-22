package smugmug

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserService is the API for user endpoints
type UserService service

func (s *UserService) expand(user *User, expansions map[string]*json.RawMessage) (*User, error) {
	if val, ok := expansions[user.URIs.Node.URI]; ok {
		res := struct{ Node *Node }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		user.Node = res.Node
	}
	if val, ok := expansions[user.URIs.Folder.URI]; ok {
		res := struct{ Folder *Folder }{}
		if err := json.Unmarshal(*val, &res); err != nil {
			return nil, err
		}
		user.Folder = res.Folder
	}
	return user, nil
}

// AuthUser returns the authorized user
func (s *UserService) AuthUser(ctx context.Context, options ...APIOption) (*User, error) {
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
