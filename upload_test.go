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

	"github.com/bzimmer/smugmug"
)

func TestUpload(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name     string
		album    string
		filename string
		err      string
	}{
		{
			name:     "no album",
			filename: "DSC33556.jpg",
			err:      "missing albumKey",
		},
		{
			name:     "with album",
			album:    "7dFHSm",
			filename: "DSC33556.jpg",
		},
		{
			name:     "filename with spaces",
			album:    "7dFHSm",
			filename: "This is a name with spaces.jpg",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			mux.HandleFunc("/photo.jpg", func(w http.ResponseWriter, r *http.Request) {
				a.Equal(http.MethodPut, r.Method)
				a.Equal(url.PathEscape(tt.filename), r.Header.Get("X-Smug-FileName"))
				http.ServeFile(w, r, "testdata/upload_CVvj69L.json")
			})
			svr := httptest.NewServer(mux)
			defer svr.Close()
			mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL), smugmug.WithUploadURL(svr.URL))
			a.NoError(err)

			up := &smugmug.Uploadable{
				Name:     tt.filename,
				AlbumKey: tt.album,
			}
			upload, err := mg.Upload.Upload(context.TODO(), up)
			if tt.err != "" {
				a.Error(err)
				a.Contains(err.Error(), tt.err)
			} else {
				a.NoError(err)
				a.NotNil(upload)
			}
		})
	}
}

type testUploadables struct {
	albumKey string
	sleep    time.Duration
}

func (t *testUploadables) Uploadables(_ context.Context) (<-chan *smugmug.Uploadable, <-chan error) {
	errc := make(chan error)
	uploadablesc := make(chan *smugmug.Uploadable, 1)

	go func() {
		defer close(errc)
		defer close(uploadablesc)
		if t.sleep > 0 {
			time.Sleep(t.sleep)
		}
		uploadablesc <- &smugmug.Uploadable{Name: "DSC33556.jpg", AlbumKey: t.albumKey}
	}()

	return uploadablesc, errc
}

func TestUploads(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		name     string
		albumKey string
		filename string
		sleep    time.Duration
		f        func(*smugmug.Upload, error)
	}{
		{
			name:     "nil album key",
			albumKey: "",
			f: func(up *smugmug.Upload, err error) {
				a.Error(err)
				a.Nil(up)
			},
		},
		{
			name:     "passing",
			albumKey: "7dFHSm",
			filename: "testdata/upload_CVvj69L.json",
			f: func(up *smugmug.Upload, err error) {
				a.Nil(err)
				a.Equal("/api/v2/image/CVvj69L-0", up.ImageURI)
			},
		},
		{
			name:     "exceed deadline",
			albumKey: "7dFHSm",
			filename: "testdata/upload_CVvj69L.json",
			sleep:    time.Minute * 1,
			f: func(up *smugmug.Upload, err error) {
				a.Nil(up)
				a.Error(err)
				a.True(errors.Is(err, context.DeadlineExceeded))
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, tt.filename)
			}))
			defer svr.Close()

			var err error
			mg, err := smugmug.NewClient(
				smugmug.WithBaseURL(svr.URL),
				smugmug.WithUploadURL(svr.URL),
				smugmug.WithConcurrency(4))
			a.NoError(err)

			ctx := context.TODO()
			if tt.sleep > 0 {
				var cancel func()
				ctx, cancel = context.WithTimeout(context.TODO(), time.Millisecond*10)
				defer cancel()
			}

			uploadables := &testUploadables{albumKey: tt.albumKey, sleep: tt.sleep}
			uploadc, errc := mg.Upload.Uploads(ctx, uploadables)
			a.NotNil(uploadc)
			a.NotNil(errc)

			err = nil
			var up *smugmug.Upload
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case err = <-errc:
			case up = <-uploadc:
			}

			tt.f(up, err)
		})
	}
}
