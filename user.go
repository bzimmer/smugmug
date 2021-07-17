package smugmug

import (
	"context"
	"net/http"
)

// UserService is the API for user endpoints
type UserService service

func (s *UserService) User(ctx context.Context) (*User, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "!authuser")
	if err != nil {
		return nil, err
	}
	res := &UserResponse{}
	err = s.client.do(req, res)
	if err != nil {
		return nil, err
	}
	return res.Response.User, err
}
