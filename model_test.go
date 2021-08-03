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

	tests := []string{
		`{"C": 4.0}`,
		`{"C": "4.0"}`,
	}

	for i := range tests {
		var qq q
		a.NoError(json.Unmarshal([]byte(tests[i]), &qq))
		a.Equal(4.0, float64(qq.C))
	}
}
