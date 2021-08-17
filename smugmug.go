package smugmug

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/armon/go-metrics"
	"github.com/mrjones/oauth"
)

//go:generate genwith --client --do --package smugmug

const (
	batchSize = 100
	userAgent = "github.com/bzimmer/smugmug"

	_baseURL   = "https://api.smugmug.com/api/v2"
	_uploadURL = "https://upload.smugmug.com"
)

// Provider specifies OAuth 1.0 URLs for SmugMug
var Provider = oauth.ServiceProvider{
	RequestTokenUrl:   "https://api.smugmug.com/services/oauth/1.0a/getRequestToken",
	AuthorizeTokenUrl: "https://api.smugmug.com/services/oauth/1.0a/authorize",
	AccessTokenUrl:    "https://api.smugmug.com/services/oauth/1.0a/getAccessToken",
}

// Client provides SmugMug connectivity
type Client struct {
	client    *http.Client
	pretty    bool
	baseURL   string
	uploadURL string
	metrics   *metrics.Metrics

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
		if c.metrics == nil {
			cfg := metrics.DefaultConfig("smugmug")
			met, err := metrics.New(cfg, &metrics.BlackholeSink{})
			if err != nil {
				return err
			}
			c.metrics = met
		}
		return nil
	}
}

// APIOption for configuring API requests
type APIOption func(url.Values) error

// WithMetrics configures the metrics instance
func WithMetrics(metrics *metrics.Metrics) Option {
	return func(c *Client) error {
		c.metrics = metrics
		return nil
	}
}

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

// WithSorting specifies sorting direction and method
// The allowable values change with the context (eg albums, nodes, folders)
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

// URLName returns `name` as a suitable URL name for a folder or album
func URLName(name string) string {
	c := regexp.MustCompile("[A-Za-z0-9-]+")
	return strings.Join(c.FindAllString(strings.Title(name), -1), "-")
}

// NewHTTPClient is a convenience function for creating an OAUTH1-compatible http client
func NewHTTPClient(consumerKey, consumerSecret, accessToken, accessTokenSecret string) (*http.Client, error) {
	consumer := oauth.NewConsumer(consumerKey, consumerSecret, Provider)
	token := &oauth.AccessToken{Token: accessToken, Secret: accessTokenSecret}
	return consumer.MakeHttpClient(token)
}

// newRequest constructs an http.Request for the uri applying all provided `APIOption`s
func (c *Client) newRequest(ctx context.Context, method, uri string, options []APIOption) (*http.Request, error) {
	return c.newRequestWithBody(ctx, method, uri, nil, options)
}

// newRequest constructs an http.Request for the uri applying all provided `APIOption`s
func (c *Client) newRequestWithBody(ctx context.Context, method, uri string, body io.Reader, options []APIOption) (*http.Request, error) {
	uri = fmt.Sprintf("%s/%s", c.baseURL, uri)
	switch method {
	case http.MethodGet:
		v := url.Values{"_pretty": {strconv.FormatBool(c.pretty)}}
		for _, opt := range options {
			if err := opt(v); err != nil {
				return nil, err
			}
		}
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	case http.MethodPost:
	default:
		return nil, fmt.Errorf("unsupported method {%s}", method)
	}
	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}
