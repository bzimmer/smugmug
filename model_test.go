package smugmug_test

import (
	"encoding/json"
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
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
		C smugmug.Coordinate
	}

	tests := []struct {
		name  string
		value string
		fail  bool
	}{
		{name: "float", value: `{"C": 4.0}`},
		{name: "string", value: `{"C": "4.0"}`},
		{name: "invalid", value: `{"C": }`, fail: true},
		{name: "not a float", value: `{"C": "abc"}`, fail: true},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var qq q
			err := json.Unmarshal([]byte(test.value), &qq)
			if !test.fail {
				a.NoError(err)
				a.Equal(4.0, float64(qq.C))
			} else {
				a.Error(err)
			}
		})
	}
}
