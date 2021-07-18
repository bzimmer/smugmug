package smugmug

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// UserService is the API for user endpoints
type UserService service

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

	if res.Expansions != nil {
		res.Response.User.Expansions = &UserExpansions{}
		for key, val := range res.Expansions {
			switch {
			case strings.HasSuffix(key, "!albums"):
				type Exp struct {
					Album []*Album
				}
				exp := &Exp{}
				err := json.Unmarshal(*val, exp)
				if err != nil {
					return nil, err
				}
				res.Response.User.Expansions.Albums = exp.Album
			}
		}
	}
	return res.Response.User, err
}
