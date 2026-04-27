package smugmug

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bzimmer/httpwares"
)

type service struct {
	client *Client
}

// Option provides a configuration mechanism for a Client.
type Option func(*Client) error

// NewClient creates a new client and applies all provided Options.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		client: &http.Client{},
	}
	opts = append(opts, withServices())
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithHTTPTracing enables tracing http calls.
func WithHTTPTracing(debug bool) Option {
	return func(c *Client) error {
		if !debug {
			return nil
		}
		c.client.Transport = &httpwares.VerboseTransport{
			Transport: c.client.Transport,
		}
		return nil
	}
}

// WithTransport sets the underlying http client transport.
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) error {
		if t == nil {
			return errors.New("nil transport")
		}
		c.client.Transport = t
		return nil
	}
}

// WithHTTPClient sets the underlying http client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("nil client")
		}
		c.client = client
		return nil
	}
}

// do executes the http request and populates v with the result.
func (c *Client) do(req *http.Request, v any) error {
	ctx := req.Context()
	res, err := c.client.Do(req) //nolint:gosec // G704: request URI is constructed by the library, not from user input
	if err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return err
		}
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusBadRequest {
		return c.decodeError(res)
	}
	return c.decode(res.Body, v)
}

// decodeError decodes an HTTP error response into a Fault.
func (c *Client) decodeError(res *http.Response) error {
	f := &Fault{}
	// json.Decoder.Decode returns io.EOF when the response body is empty (e.g. some 4xx responses).
	if err := json.NewDecoder(res.Body).Decode(f); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if f.Code == 0 {
		f.Code = res.StatusCode
	}
	if f.Message == "" {
		f.Message = http.StatusText(res.StatusCode)
	}
	return f
}

// decode decodes the response body into v.
// json.Decoder.Decode returns io.EOF for empty response bodies (e.g. DELETE responses), which is ignored.
func (c *Client) decode(body io.Reader, v any) error {
	if v == nil {
		return nil
	}
	if err := json.NewDecoder(body).Decode(v); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
