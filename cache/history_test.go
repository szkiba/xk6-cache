package cache

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistory_put(t *testing.T) {
	t.Parallel()

	var cache history

	loc, _ := url.Parse("https://example.com?foo=bar")

	cache.put(loc, &reply{header: nil, body: nil})

	_, found := cache.get(loc)

	assert.True(t, found)

	from := &reply{header: http.Header{"Foo": []string{"Bar"}}, body: []byte("Hello World!")}

	cache.put(loc, from)

	rep, found := cache.get(loc)

	assert.True(t, found)
	assert.NotNil(t, rep)
	assert.Contains(t, rep.header, "Content-Location")
	assert.Equal(t, "https://example.com?foo=bar", rep.header.Get("Content-Location"))
	assert.Contains(t, rep.header, "Content-Disposition")
	assert.Equal(t, `attachment; filename="https://example.com"`, rep.header.Get("Content-Disposition"))
	assert.Contains(t, rep.header, "Content-Length")
	assert.Equal(t, strconv.Itoa(len("Hello World!")), rep.header.Get("Content-Length"))
}

func TestHistory_get(t *testing.T) {
	t.Parallel()

	var cache history

	loc, _ := url.Parse("https://example.com")

	_, found := cache.get(loc)

	assert.False(t, found)
}

func TestHistory_marshalHeader(t *testing.T) {
	t.Parallel()

	var cache history

	var buff bytes.Buffer

	assert.NoError(t, cache.marshalHeader(&buff))
	assert.Equal(t, "Content-Type: multipart/mixed; boundary=______________________________o_o______________________________\r\nSubject: xk6-cache\r\n\r\n", buff.String())
}

func TestHistory_marshalError(t *testing.T) {
	t.Parallel()

	var cache history

	file, _ := os.Open("history_test.go")

	assert.Error(t, cache.marshalHeader(file))
	assert.Error(t, cache.marshal(file))
}

func TestHistory_unmarshalError(t *testing.T) {
	t.Parallel()

	var cache history

	content := ``

	_, _, err := cache.unmarshalHeader(strings.NewReader(content))

	assert.Error(t, err)
	assert.Error(t, cache.unmarshal(strings.NewReader(content)))
}

func TestHistory_marshal(t *testing.T) {
	t.Parallel()

	var cache history

	rep := &reply{header: nil, body: []byte("Hello .net World!")}
	loc, _ := url.Parse("https://example.net")

	cache.put(loc, rep)

	rep = &reply{header: nil, body: []byte("Hello .com World!")}
	loc, _ = url.Parse("https://example.com")

	cache.put(loc, rep)

	var buff bytes.Buffer

	assert.NoError(t, cache.marshal(&buff))

	content := buff.Bytes()

	var other history

	assert.NoError(t, other.unmarshal(&buff))
	assert.Equal(t, cache.store, other.store)

	netIndex := bytes.Index(content, []byte("Content-Location: https://example.net"))
	comIndex := bytes.Index(content, []byte("Content-Location: https://example.com"))

	assert.Greater(t, netIndex, 0)
	assert.Greater(t, comIndex, 0)
	assert.Less(t, comIndex, netIndex)
}
