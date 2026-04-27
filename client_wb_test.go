package smugmug

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeNilV(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	c := &Client{}
	err := c.decode(strings.NewReader(`{"foo":"bar"}`), nil)
	a.NoError(err)
}
