package smugmug_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
)

var errFail = errors.New("fail")

func withError() smugmug.APIOption {
	return func(v url.Values) error {
		return errFail
	}
}

func TestAPIOptions(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	opts := []smugmug.APIOption{
		smugmug.WithExpansions("Image", "Album"),
		smugmug.WithFilters("Name"),
		smugmug.WithSorting("Ascending", "LastUpdated"),
		smugmug.WithSearch("/api/v2/user/cmac", "Marmot"),
	}

	v := url.Values{}
	for i := range opts {
		a.NoError(opts[i](v))
	}

	a.Equal("Name", v.Get("_filter"))
	a.Equal("Image,Album", v.Get("_expand"))
	a.Equal("Ascending", v.Get("SortDirection"))
	a.Equal("LastUpdated", v.Get("SortMethod"))
	a.Equal("/api/v2/user/cmac", v.Get("Scope"))
	a.Equal("Marmot", v.Get("Text"))

	a.NoError(smugmug.WithSorting("", "LastUploaded")(v))
	a.NoError(smugmug.WithSearch("", "Marmot")(v))
	a.Equal("", v.Get("SortDirection"))
	a.Equal("LastUploaded", v.Get("SortMethod"))
	a.Equal("", v.Get("Scope"))
	a.Equal("Marmot", v.Get("Text"))
}

func TestOption(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	client, err := smugmug.NewClient(
		smugmug.WithHTTPTracing(true),
		smugmug.WithHTTPClient(http.DefaultClient),
		smugmug.WithTransport(http.DefaultTransport),
	)
	a.NoError(err)
	a.NotNil(client)

	client, err = smugmug.NewClient(smugmug.WithHTTPClient(nil))
	a.Error(err)
	a.Nil(client)

	client, err = smugmug.NewClient(smugmug.WithTransport(nil))
	a.Error(err)
	a.Nil(client)

	client, err = smugmug.NewClient(smugmug.WithHTTPTracing(false))
	a.NoError(err)
	a.NotNil(client)
}

type errorTransport struct{}

func (t *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("boom")
}

func TestDo(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/user_cmac.json")
	}))
	defer svr.Close()

	sleeper := func(dur time.Duration) smugmug.APIOption {
		return func(v url.Values) error {
			time.Sleep(dur)
			return nil
		}
	}

	client, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL))
	a.NoError(err)
	a.NotNil(client)

	ctx := context.TODO()
	user, err := client.User.AuthUser(ctx, sleeper(time.Millisecond*100))
	a.NotNil(user)
	a.NoError(err)

	client, err = smugmug.NewClient(smugmug.WithTransport(&errorTransport{}))
	a.NoError(err)
	a.NotNil(client)
	user, err = client.User.AuthUser(ctx)
	a.Nil(user)
	a.Error(err)
	a.Equal("boom", err.(*url.Error).Unwrap().Error())

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*5)
	defer cancel()
	user, err = client.User.AuthUser(ctx, sleeper(time.Millisecond*150))
	a.Nil(user)
	a.Error(err)
	a.True(errors.Is(err, context.DeadlineExceeded))
}

func TestOAuthClient(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	c, err := smugmug.NewHTTPClient("consumerKey", "consumerSecret", "accessToken", "accessTokenSecret")
	a.NoError(err)
	a.NotNil(c)
}

func TestURLName(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	a.Equal("", smugmug.URLName(""))
	a.Equal("Foo-Bar", smugmug.URLName("foo bar"))
	a.Equal("Foo-Bar", smugmug.URLName("foo bar "))
	a.Equal("Foo-Bar", smugmug.URLName("Foo bar"))
	a.Equal("Foo-1-Bar", smugmug.URLName("foo & 1 bar"))
	a.Equal("2021-01-01-Foo-1-Bar", smugmug.URLName("2021-01-01 foo & 1 bar"))
}
