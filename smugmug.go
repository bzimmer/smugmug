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
	batchSize = 100
	baseURL   = "https://api.smugmug.com/api/v2"
	uploadURL = "https://upload.smugmug.com"
	userAgent = "github.com/bzimmer/smugmug"
)

var Provider = oauth.ServiceProvider{
	RequestTokenUrl:   "https://api.smugmug.com/services/oauth/1.0a/getRequestToken",
	AuthorizeTokenUrl: "https://api.smugmug.com/services/oauth/1.0a/authorize",
	AccessTokenUrl:    "https://api.smugmug.com/services/oauth/1.0a/getAccessToken",
}

type Client struct {
	client *http.Client
	pretty bool

	User   *UserService
	Node   *NodeService
	Album  *AlbumService
	Image  *ImageService
	Upload *UploadService
}

func withServices() Option {
	return func(c *Client) error {
		c.User = &UserService{c}
		c.Node = &NodeService{c}
		c.Album = &AlbumService{c}
		c.Image = &ImageService{c}
		c.Upload = &UploadService{c}
		return nil
	}
}

// APIOption for configuring API requests
type APIOption func(url.Values) error

// WithPretty enable indention of the req/res from SmugMug (useful for debugging)
func WithPretty(pretty bool) Option {
	return func(c *Client) error {
		c.pretty = pretty
		return nil
	}
}

// WithExpansions requests expansions for single-entity queries
func WithExpansions(expansions ...string) APIOption {
	return func(v url.Values) error {
		v.Set("_expand", strings.Join(expansions, ","))
		return nil
	}
}

// WithPagination enables paging results for albums, nodes, and images
func WithPagination(start, count int) APIOption {
	return func(v url.Values) error {
		v.Set("start", fmt.Sprintf("%d", start))
		v.Set("count", fmt.Sprintf("%d", count))
		return nil
	}
}

func WithSorting(direction, method string) APIOption {
	return func(v url.Values) error {
		v.Set("SortDirection", direction)
		v.Set("SortMethod", method)
		return nil
	}
}

func WithFilters(filters ...string) APIOption {
	return func(v url.Values) error {
		v.Set("_filter", strings.Join(filters, ","))
		return nil
	}
}

// WithSearch queries Smugmug for the text within the given scope
// The scope is a URI representing a user, album, node, or folder
func WithSearch(scope, text string) APIOption {
	return func(v url.Values) error {
		v.Set("Text", text)
		v.Set("Scope", scope)
		return nil
	}
}

// NewHTTPClient is a convenience function for creating an OAUTH1-compatible http client
func NewHTTPClient(consumerKey, consumerSecret, accessToken, accessTokenSecret string) (*http.Client, error) {
	consumer := oauth.NewConsumer(consumerKey, consumerSecret, Provider)
	token := &oauth.AccessToken{Token: accessToken, Secret: accessTokenSecret}
	return consumer.MakeHttpClient(token)
}

func (c *Client) newRequest(ctx context.Context, method, uri string, options []APIOption) (*http.Request, error) {
	if strings.HasPrefix("!", uri) {
		uri = fmt.Sprintf("%s%s", baseURL, uri)
	} else {
		uri = fmt.Sprintf("%s/%s", baseURL, uri)
	}

	v := url.Values{}
	if c.pretty {
		v.Set("_pretty", "true")
	}
	for _, opt := range options {
		if err := opt(v); err != nil {
			return nil, err
		}
	}

	if len(v) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}
