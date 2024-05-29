package smugmug_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/smugmug"
)

func TestFault(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	f := &smugmug.Fault{Message: "foo"}
	a.Error(f)
	a.Equal("foo", f.Error())
}

func TestCoordinate(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	type q struct {
		C smugmug.Coordinate `json:"C"`
	}

	tests := []struct {
		name  string
		value string
		f     func(c smugmug.Coordinate, err error)
	}{
		{
			name:  "float",
			value: `{"C": 4.0}`,
			f: func(c smugmug.Coordinate, err error) {
				a.NoError(err)
				a.Equal(4.0, float64(c))
			},
		},
		{
			name:  "string",
			value: `{"C": "4.0"}`,
			f: func(c smugmug.Coordinate, err error) {
				a.NoError(err)
				a.Equal(4.0, float64(c))
			},
		},
		{
			name:  "invalid",
			value: `{"C": }`,
			f: func(_ smugmug.Coordinate, err error) {
				a.Error(err)
			},
		},
		{
			name:  "not a float",
			value: `{"C": "abc"}`,
			f: func(_ smugmug.Coordinate, err error) {
				a.Error(err)
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var qq q
			t.Parallel()
			err := json.Unmarshal([]byte(test.value), &qq)
			test.f(qq.C, err)
		})
	}
}
