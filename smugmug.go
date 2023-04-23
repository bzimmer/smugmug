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

	"github.com/mrjones/oauth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:generate genwith --client --do --package smugmug

const (
	batch       = 100
	concurrency = 2

	userAgent = "github.com/bzimmer/smugmug"

	_baseURL   = "https://api.smugmug.com/api/v2"
	_uploadURL = "https://upload.smugmug.com"

	TypeAlbum  = "Album"
	TypeFolder = "Folder"
)

var (
	// albumNameRE allowable characters
	albumNameRE = regexp.MustCompile(`[\p{L}\d]+`)
)

// provider specifies OAuth 1.0 URLs for SmugMug
func provider() oauth.ServiceProvider {
	return oauth.ServiceProvider{
		RequestTokenUrl:   "https://api.smugmug.com/services/oauth/1.0a/getRequestToken",
		AuthorizeTokenUrl: "https://api.smugmug.com/services/oauth/1.0a/authorize",
		AccessTokenUrl:    "https://api.smugmug.com/services/oauth/1.0a/getAccessToken",
	}
}

// Client provides SmugMug connectivity
type Client struct {
	client      *http.Client
	pretty      bool
	baseURL     string
	uploadURL   string
	concurrency int

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
		if c.concurrency == 0 {
			c.concurrency = concurrency
		}
		return nil
	}
}

// APIOption for configuring API requests
type APIOption func(url.Values) error

// WithConcurrency configures the number of concurrent upload goroutines
func WithConcurrency(concurrency int) Option {
	return func(c *Client) error {
		c.concurrency = concurrency
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

// searchReplaceREs special characters to replace
func searchReplaceREs() map[*regexp.Regexp]string {
	re := make(map[*regexp.Regexp]string, 0)
	for search, replace := range map[string]string{
		`-`:                    " ",
		"[" + "`" + `'"` + "]": "",
	} {
		c := regexp.MustCompile(search)
		re[c] = replace
	}
	return re
}

// URLName returns `name` as a suitable URL name for a folder or album
func URLName(name string, tags ...language.Tag) string {
	s := name
	tag := language.Und
	if len(tags) > 0 {
		tag = tags[0]
	}
	upper := cases.Upper(tag)
	for search, replace := range searchReplaceREs() {
		s = search.ReplaceAllString(s, replace)
	}
	t := albumNameRE.FindAllString(s, -1)
	for i := 0; i < len(t); i++ {
		u := t[i]
		// if the part is entirely capitals, probably an acronym
		if upper.String(u) != u {
			t[i] = cases.Title(tag).String(u)
		}
	}
	return strings.Join(t, "-")
}

// NewHTTPClient is a convenience function for creating an OAUTH1-compatible http client
func NewHTTPClient(consumerKey, consumerSecret, accessToken, accessTokenSecret string) (*http.Client, error) {
	consumer := oauth.NewConsumer(consumerKey, consumerSecret, provider())
	token := &oauth.AccessToken{Token: accessToken, Secret: accessTokenSecret}
	return consumer.MakeHttpClient(token)
}

// newRequest constructs an http.Request for the uri applying all provided `APIOption`s
func (c *Client) newRequest(ctx context.Context, uri string, options []APIOption) (*http.Request, error) {
	return c.newRequestWithBody(ctx, http.MethodGet, uri, nil, options)
}

// newRequest constructs an http.Request for the uri applying all provided `APIOption`s
func (c *Client) newRequestWithBody(
	ctx context.Context, method, uri string, body io.Reader, options []APIOption) (*http.Request, error) {
	uri = fmt.Sprintf("%s/%s", c.baseURL, uri)
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
		v := url.Values{"_pretty": {strconv.FormatBool(c.pretty)}}
		for _, opt := range options {
			if err := opt(v); err != nil {
				return nil, err
			}
		}
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
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

func iterate[T any](ctx context.Context,
	q func(ctx context.Context, options ...APIOption) ([]T, *Pages, error),
	f func(T) (bool, error), options ...APIOption) error {
	var n int
	page := WithPagination(1, batch)
	for {
		nodes, pages, err := q(ctx, append(options, page)...)
		if err != nil {
			return err
		}
		n += pages.Count
		for _, node := range nodes {
			var ok bool
			if ok, err = f(node); err != nil {
				return err
			} else if !ok {
				return nil
			}
		}
		if n == pages.Total {
			return nil
		}
		page = WithPagination(pages.Start+pages.Count, batch)
	}
}
