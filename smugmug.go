package smugmug

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrjones/oauth"
)

//go:generate genwith --client --do --package smugmug

const (
	baseURL   = "https://api.smugmug.com/api/v2"
	userAgent = "github.com/bzimmer/smugmug"
)

var Provider = oauth.ServiceProvider{
	RequestTokenUrl:   "https://api.smugmug.com/services/oauth/1.0a/getRequestToken",
	AuthorizeTokenUrl: "https://api.smugmug.com/services/oauth/1.0a/authorize",
	AccessTokenUrl:    "https://api.smugmug.com/services/oauth/1.0a/getAccessToken",
}

type Client struct {
	client *http.Client

	User  *UserService
	Album *AlbumService
	Image *ImageService
}

func withServices() Option {
	return func(c *Client) error {
		c.User = &UserService{c}
		c.Album = &AlbumService{c}
		c.Image = &ImageService{c}
		return nil
	}
}

func NewHTTPClient(consumerKey, consumerSecret, accessToken, accessTokenSecret string) (*http.Client, error) {
	consumer := oauth.NewConsumer(consumerKey, consumerSecret, Provider)
	token := &oauth.AccessToken{Token: accessToken, Secret: accessTokenSecret}
	return consumer.MakeHttpClient(token)
}

func (c *Client) newRequest(ctx context.Context, method, uri string) (*http.Request, error) {
	if strings.HasPrefix("!", uri) {
		uri = fmt.Sprintf("%s%s", baseURL, uri)
	} else {
		uri = fmt.Sprintf("%s/%s", baseURL, uri)
	}
	fmt.Println("< " + uri)
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	fmt.Println("> " + uri)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}
