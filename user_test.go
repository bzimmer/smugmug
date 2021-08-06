package smugmug_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
)

func TestAuthUser(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name     string
		filename string
		options  []smugmug.APIOption
		f        func(user *smugmug.User, err error)
	}{
		{
			name:     "success",
			filename: "testdata/user_cmac.json",
			f: func(user *smugmug.User, err error) {
				a.NoError(err)
				a.NotNil(user)
			},
		},
		{
			name:     "failed parse",
			filename: "user_test.go",
			f: func(user *smugmug.User, err error) {
				a.Error(err)
				a.Nil(user)
			},
		},
		{
			name:     "failed options",
			filename: "testdata/user_cmac.json",
			options:  []smugmug.APIOption{withError(true)},
			f: func(user *smugmug.User, err error) {
				a.Error(err)
				a.True(errors.Is(err, withErr))
				a.Nil(user)
			},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fp, err := os.Open(test.filename)
				a.NoError(err)
				defer fp.Close()
				_, err = io.Copy(w, fp)
				a.NoError(err)
			}))
			defer svr.Close()

			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL), smugmug.WithPretty(false))
			a.NoError(err)
			user, err := mg.User.AuthUser(context.Background(), test.options...)
			test.f(user, err)
		})
	}
}
