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

func TestISO(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// Direct call with invalid bytes exercises the JSON-parse error path.
	var iso smugmug.ISO
	a.Error(iso.UnmarshalJSON([]byte(`not-valid-json`)))

	type q struct {
		I smugmug.ISO `json:"I"`
	}

	tests := []struct {
		name  string
		value string
		f     func(i smugmug.ISO, err error)
	}{
		{
			name:  "float",
			value: `{"I": 100.0}`,
			f: func(i smugmug.ISO, err error) {
				a.NoError(err)
				a.Equal(smugmug.ISO(100), i)
			},
		},
		{
			name:  "string numeric",
			value: `{"I": "200"}`,
			f: func(i smugmug.ISO, err error) {
				a.NoError(err)
				a.Equal(smugmug.ISO(200), i)
			},
		},
		{
			name:  "string non-numeric falls back to zero",
			value: `{"I": "auto"}`,
			f: func(i smugmug.ISO, err error) {
				a.NoError(err)
				a.Equal(smugmug.ISO(0), i)
			},
		},
		{
			name:  "null (no-op)",
			value: `{"I": null}`,
			f: func(i smugmug.ISO, err error) {
				a.NoError(err)
				a.Equal(smugmug.ISO(0), i)
			},
		},
		{
			name:  "invalid json",
			value: `{"I": }`,
			f: func(_ smugmug.ISO, err error) {
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
			test.f(qq.I, err)
		})
	}
}

func TestCoordinate(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// Direct call to UnmarshalJSON with invalid bytes (not valid JSON).
	var c smugmug.Coordinate
	a.Error(c.UnmarshalJSON([]byte(`not-valid-json`)))

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
		{
			name:  "null (no-op)",
			value: `{"C": null}`,
			f: func(c smugmug.Coordinate, err error) {
				a.NoError(err)
				a.Equal(0.0, float64(c))
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

func TestAltitude(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// Direct call to UnmarshalJSON with invalid bytes (not valid JSON).
	var alt smugmug.Altitude
	a.Error(alt.UnmarshalJSON([]byte(`not-valid-json`)))

	type q struct {
		A smugmug.Altitude `json:"A"`
	}

	tests := []struct {
		name  string
		value string
		f     func(c smugmug.Altitude, err error)
	}{
		{
			name:  "float",
			value: `{"A": 4.0}`,
			f: func(c smugmug.Altitude, err error) {
				a.NoError(err)
				a.Equal("4.000000", string(c))
			},
		},
		{
			name:  "string m",
			value: `{"A": "168.1 m"}`,
			f: func(c smugmug.Altitude, err error) {
				a.NoError(err)
				a.Equal("168.1 m", string(c))
			},
		},
		{
			name:  "string m above sea level",
			value: `{"A": "1700.2 m Above Sea Level"}`,
			f: func(c smugmug.Altitude, err error) {
				a.NoError(err)
				a.Equal("1700.2 m Above Sea Level", string(c))
			},
		},
		{
			name:  "null (no-op)",
			value: `{"A": null}`,
			f: func(c smugmug.Altitude, err error) {
				a.NoError(err)
				a.Equal("", string(c))
			},
		},
		{
			name:  "invalid json",
			value: `{"A": }`,
			f: func(_ smugmug.Altitude, err error) {
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
			test.f(qq.A, err)
		})
	}
}
