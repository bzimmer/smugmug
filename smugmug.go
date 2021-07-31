package smugmug

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mrjones/oauth"
)

//go:generate genwith --client --do --package smugmug

const (
	batchSize = 100
	userAgent = "github.com/bzimmer/smugmug"

	_baseURL   = "https://api.smugmug.com/api/v2"
	_uploadURL = "https://upload.smugmug.com"
)

var Provider = oauth.ServiceProvider{
	RequestTokenUrl:   "https://api.smugmug.com/services/oauth/1.0a/getRequestToken",
	AuthorizeTokenUrl: "https://api.smugmug.com/services/oauth/1.0a/authorize",
	AccessTokenUrl:    "https://api.smugmug.com/services/oauth/1.0a/getAccessToken",
}

type Client struct {
	client    *http.Client
	pretty    bool
	baseURL   string
	uploadURL string

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

		if c.baseURL == "" {
			c.baseURL = _baseURL
		}
		if c.uploadURL == "" {
			c.uploadURL = _uploadURL
		}
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

// WithBaseURL specifies the base url
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		c.baseURL = baseURL
		return nil
	}
}

// WithUploadURL specifies the upload url
func WithUploadURL(uploadURL string) Option {
	return func(c *Client) error {
		c.uploadURL = uploadURL
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
		v.Del("SortDirection")
		if direction != "" {
			v.Set("SortDirection", direction)
		}
		v.Del("SortMethod")
		if method != "" {
			v.Set("SortMethod", method)
		}
		return nil
	}
}

// WithFilters queries SmugMug for only the attributes in the filter list
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
		v.Del("Text")
		if text != "" {
			v.Set("Text", text)
		}
		v.Del("Scope")
		if scope != "" {
			v.Set("Scope", scope)
		}
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
		uri = fmt.Sprintf("%s%s", c.baseURL, uri)
	} else {
		uri = fmt.Sprintf("%s/%s", c.baseURL, uri)
	}

	v := url.Values{}
	v.Set("_pretty", strconv.FormatBool(c.pretty))

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
