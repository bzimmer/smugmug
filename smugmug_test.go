package smugmug_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
)

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
}
