package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_response2reply(t *testing.T) {
	t.Parallel()

	body := []byte("Hello World!")
	hdr := http.Header{"Access-Control-Allow-Origin": []string{"*"}}

	from := &http.Response{ // nolint:exhaustruct
		Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
		ContentLength: int64(len(body)),
		Header:        hdr,
	}

	rep, err := response2reply(from)

	assert.NoError(t, err)
	assert.NotNil(t, rep)

	assert.Equal(t, body, rep.body)
	assert.Equal(t, hdr, rep.header)

	content, err := io.ReadAll(from.Body)

	assert.NoError(t, err)

	assert.Equal(t, body, content)

	file, err := os.Open("tripperware_test.go")

	assert.NoError(t, err)

	assert.NoError(t, file.Close())

	from.Body = file

	_, err = response2reply(from)

	assert.Error(t, err)
}

func Test_reply2respose(t *testing.T) {
	t.Parallel()

	from := &reply{header: http.Header{"Foo": []string{"Bar"}}, body: []byte("Hello World!")}
	req := new(http.Request)
	res := reply2response(req, from)

	defer res.Body.Close()

	assert.NotNil(t, res)
	assert.Same(t, req, res.Request)
	assert.Equal(t, http.StatusText(http.StatusOK), res.Status)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "HTTP/1.1", res.Proto)
	assert.Equal(t, 1, res.ProtoMajor)
	assert.Equal(t, 1, res.ProtoMinor)
	assert.Equal(t, from.header, res.Header)
	assert.Equal(t, int64(len(from.body)), res.ContentLength)

	body, err := io.ReadAll(res.Body)

	assert.NoError(t, err)
	assert.Equal(t, from.body, body)
}

func TestTripperware_saveError(t *testing.T) {
	t.Parallel()

	from := new(tripperware)

	from.history = new(history)
	from.logger = logrus.StandardLogger()

	err := from.save("")

	assert.Error(t, err)
}

func TestTripperware_save(t *testing.T) {
	t.Parallel()

	from := new(tripperware)

	from.history = new(history)
	from.logger = logrus.StandardLogger()

	loc, _ := url.Parse("https://example.com?foo=bar")

	rep := &reply{header: http.Header{"Foo": []string{"Bar"}}, body: []byte("Hello World!")}

	from.history.put(loc, rep)

	file, err := ioutil.TempFile("", "")

	assert.NoError(t, err)

	err = from.save(file.Name())

	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	file, err = os.Open(file.Name())

	assert.NoError(t, err)

	saved := new(history)

	assert.NoError(t, saved.unmarshal(file))
	assert.Equal(t, from.history, saved)
}

func TestTripperware_shouldStore(t *testing.T) {
	t.Parallel()

	tw := new(tripperware) // nolint:varnamelen
	res := new(http.Response)

	res.StatusCode = http.StatusBadRequest

	assert.False(t, tw.shouldStore(res))

	res.StatusCode = http.StatusOK

	assert.True(t, tw.shouldStore(res))

	res.Header = http.Header{}

	res.Header.Set("Content-Type", "application/octet-stream")

	assert.False(t, tw.shouldStore(res))

	res.Header.Set("Content-Type", "application/javascript")

	assert.True(t, tw.shouldStore(res))

	res.Header.Set("Content-Type", "text/javascript")

	assert.True(t, tw.shouldStore(res))

	res.Header.Set("Content-Type", "text/javascript;charset=UTF-8")

	assert.True(t, tw.shouldStore(res))

	res.Header.Set("Content-Type", "text/plain")

	assert.True(t, tw.shouldStore(res))
}

func TestTripperware_RoundTrip_miss(t *testing.T) {
	t.Parallel()

	transport := newTransport(t)

	tw := newTripperware(transport, logrus.StandardLogger()) // nolint:varnamelen

	assert.NotNil(t, tw.history)

	req := new(http.Request)

	loc, _ := url.Parse("https://example.com")

	req.URL = loc

	res, err := tw.RoundTrip(req) // nolint:bodyclose

	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.NotEmpty(t, tw.history.store)
}

func TestTripperware_RoundTrip_hit(t *testing.T) {
	t.Parallel()

	transport := newTransport(t)

	tw := newTripperware(transport, logrus.StandardLogger()) // nolint:varnamelen

	assert.NotNil(t, tw.history)

	req := new(http.Request)

	loc, _ := url.Parse("https://example.com?_k6=1")

	req.URL = loc

	res, err := tw.RoundTrip(req) // nolint:bodyclose

	assert.NoError(t, err)
	assert.NotNil(t, res)

	historySize := len(tw.history.store)

	res, err = tw.RoundTrip(req) // nolint:bodyclose

	assert.NoError(t, err)
	assert.NotNil(t, res)

	logrus.Error(t, tw.history.store)
	assert.Equal(t, historySize, len(tw.history.store))
}

type testTransport struct{}

func (tt *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res := new(http.Response)

	res.Body = io.NopCloser(strings.NewReader(`Hello World!`))
	res.Request = req
	res.StatusCode = http.StatusOK

	return res, nil
}

func newTransport(t *testing.T) *testTransport {
	t.Helper()

	tt := &testTransport{}

	return tt
}
