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
	return user, nil
}

// AuthUser returns the authorized user
func (s *UserService) AuthUser(ctx context.Context, options ...APIOption) (*User, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "!authuser", options)
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
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		User           *User   `json:"User"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		DocURI         string  `json:"DocUri"`
		Timing         *timing `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}
