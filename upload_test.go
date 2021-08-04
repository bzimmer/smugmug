package smugmug_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
)

func TestUpload(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp, err := os.Open("testdata/upload_CVvj69L.json")
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL), smugmug.WithUploadURL(svr.URL))
	a.NoError(err)

	up := &smugmug.Uploadable{Name: "DSC33556.jpg"}
	upload, err := mg.Upload.Upload(context.Background(), up)
	a.Error(err)
	a.Nil(upload)

	up.AlbumKey = "7dFHSm"
	upload, err = mg.Upload.Upload(context.Background(), up)
	a.NoError(err)
	a.NotNil(upload)
}

type testUploadables struct {
	albumKey string
}

func (t *testUploadables) Uploadables(ctx context.Context) (<-chan *smugmug.Uploadable, <-chan error) {
	errs := make(chan error)
	uploadables := make(chan *smugmug.Uploadable, 1)
	uploadables <- &smugmug.Uploadable{Name: "DSC33556.jpg", AlbumKey: t.albumKey}

	close(errs)
	close(uploadables)

	return uploadables, errs
}

func TestUploads(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tests := []struct {
		albumKey string
		fail     bool
	}{
		{albumKey: "", fail: true},
		{albumKey: "7dFHSm", fail: false},
	}

	for i := range tests {
		test := tests[i]
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fp, err := os.Open("testdata/upload_CVvj69L.json")
			a.NoError(err)
			defer fp.Close()
			_, err = io.Copy(w, fp)
			a.NoError(err)
		}))
		defer svr.Close()

		mg, err := smugmug.NewClient(smugmug.WithBaseURL(svr.URL), smugmug.WithUploadURL(svr.URL))
		a.NoError(err)

		uploadables := &testUploadables{test.albumKey}
		uploadc, errc := mg.Upload.Uploads(context.Background(), uploadables)
		a.NotNil(uploadc)
		a.NotNil(errc)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		err = nil
		var up *smugmug.Upload
		select {
		case <-ctx.Done():
			a.Error(ctx.Err())
		case err = <-errc:
		case up = <-uploadc:
		}

		switch {
		case test.fail:
			a.NotNil(err)
		case !test.fail:
			a.Nil(err)
			a.NotNil(up)
			a.Equal("/api/v2/image/CVvj69L-0", up.UploadedImage.ImageURI)
		}
	}
}
