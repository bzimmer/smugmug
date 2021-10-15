package smugmug

import (
	"context"
	"encoding/json"
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
	return user, nil
}

// AuthUser returns the authorized user
func (s *UserService) AuthUser(ctx context.Context, options ...APIOption) (*User, error) {
	req, err := s.client.newRequest(ctx, "!authuser", options)
	if err != nil {
		return nil, err
	}
	res := &userResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return s.expand(res.Response.User, res.Expansions)
}

type userResponse struct {
	Response struct {
		User *User `json:"User"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}
