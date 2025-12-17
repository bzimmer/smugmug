package smugmug_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
			options:  []smugmug.APIOption{withError()},
			f: func(user *smugmug.User, err error) {
				a.Error(err)
				a.ErrorIs(err, errFail)
				a.Nil(user)
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				a.Equal("/!authuser", r.URL.Path)
				http.ServeFile(w, r, tt.filename)
			}))
			defer svr.Close()

			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL), smugmug.WithPretty(false))
			a.NoError(err)
			user, err := mg.User.AuthUser(context.TODO(), tt.options...)
			tt.f(user, err)
		})
	}
}
