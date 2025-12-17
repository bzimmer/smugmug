package smugmug_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/bzimmer/smugmug"
)

var errFail = errors.New("fail")

func withError() smugmug.APIOption {
	return func(_ url.Values) error {
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
		return func(_ url.Values) error {
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
	var q *url.Error
	ok := errors.As(err, &q)
	a.True(ok)
	a.Equal("boom", q.Unwrap().Error())

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*5)
	defer cancel()
	user, err = client.User.AuthUser(ctx, sleeper(time.Millisecond*150))
	a.Nil(user)
	a.Error(err)
	a.ErrorIs(err, context.DeadlineExceeded)
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

	tests := []struct {
		url   string
		album string
		lang  language.Tag
	}{
		{url: "", album: ""},
		{url: "Foo-Bar", album: "foo bar"},
		{url: "Foo-Bar", album: "foo bar "},
		{url: "Foo-Bar", album: "Foo bar"},
		{url: "Foo-1-Bar", album: "foo & 1 bar"},
		{url: "Someones-Something", album: "Someone's something"},
		{url: "Someones-Something", album: `Someone"s something`},
		{url: "2022-03-04-Zürich", album: "2022-03-04 Zürich"},
		{url: "2022-03-04-Zürich", album: "2022-03-04 Zürich", lang: language.German},
		{url: "2022-03-04-Zürich", album: "2022-03-04-Zürich-", lang: language.German},
		{url: "2022-03-04-Zürich", album: "2022-03-04 Zürich & ___", lang: language.German},
		{url: "2022-03-04-Zürich", album: "2022-03-04 Zürich & ___"},
		{url: "2021-01-01-Foo-1-Bar", album: "2021-01-01 foo & 1 bar"},
		{url: "Foo-1-Bar", album: "foo & 1 bar", lang: language.English},
		{url: "2009-10-11-BIFD-Pancake-Breakfast",
			album: "2009-10-11 BIFD Pancake Breakfast", lang: language.English},
		{url: "2009-10-11-BIFD-Pancake-Breakfast",
			album: "2009-10-11 BIFD `Pancake Breakfast`", lang: language.English},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.url+test.album, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			switch test.lang.IsRoot() {
			case true:
				a.Equal(test.url, smugmug.URLName(test.album))
			case false:
				a.Equal(test.url, smugmug.URLName(test.album, test.lang))
			}
		})
	}
}
