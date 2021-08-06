package smugmug_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
)

var withErr = errors.New("fail")

func withError(fail bool) smugmug.APIOption {
	return func(v url.Values) error {
		if fail {
			return withErr
		}
		return nil
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
		fp, err := os.Open("testdata/user_cmac.json")
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
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

	ctx := context.Background()
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
