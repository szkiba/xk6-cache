package cache

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneHeader(t *testing.T) {
	t.Parallel()

	from := http.Header{
		"Foo": []string{"bar", "dummy"},
	}

	to := cloneHeader(from)

	assert.Equal(t, from, to)
	assert.NotSame(t, from, to)
	assert.NotEmpty(t, to["Foo"])
	assert.NotSame(t, from["Foo"], to["Foo"])
}

func TestFilterHeader(t *testing.T) {
	t.Parallel()

	from := http.Header{
		"Content-Length":              []string{"42"},
		"Content-Type":                []string{"text/plain"},
		"Location":                    []string{"https://example.com"},
		"Access-Control-Allow-Origin": []string{"*"},
	}

	to := filterHeader(from) // nolint:varnamelen

	assert.NotSame(t, from, to)
	assert.NotEqual(t, from, to)
	assert.Contains(t, to, "Content-Length")
	assert.Contains(t, to, "Content-Type")
	assert.Contains(t, to, "Access-Control-Allow-Origin")
	assert.NotContains(t, to, "Location")
}

func TestDuplicateBody(t *testing.T) {
	t.Parallel()

	body := []byte("Hello World!")

	from := &http.Response{ // nolint:exhaustruct
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
		ContentLength: int64(len(body)),
		Request:       nil,
		Header:        http.Header{},
	}

	to, err := duplicateBody(from)

	assert.NoError(t, err)

	assert.Equal(t, body, to)
}

func TestAddK6QueryParam(t *testing.T) {
	t.Parallel()

	loc, err := url.Parse("https://example.com")

	assert.NoError(t, err)
	addK6QueryParam(loc)

	assert.Contains(t, loc.Query(), "_k6")
	assert.Equal(t, loc.Query(), url.Values{"_k6": []string{"1"}})

	loc, err = url.Parse("https://example.com?_k6=1")

	assert.NoError(t, err)
	addK6QueryParam(loc)

	assert.Contains(t, loc.Query(), "_k6")
	assert.Equal(t, loc.Query(), url.Values{"_k6": []string{"1"}})

	loc, err = url.Parse("https://example.com?foo=bar")

	assert.NoError(t, err)
	addK6QueryParam(loc)

	assert.Contains(t, loc.Query(), "_k6")
	assert.Equal(t, loc.Query(), url.Values{"_k6": []string{"1"}, "foo": []string{"bar"}})
}
