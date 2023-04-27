package cache

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewModule(t *testing.T) {
	t.Parallel()

	transport := newTransport(t)
	module := newModule("", transport, logrus.StandardLogger())

	assert.Nil(t, module.tripperware)
	assert.NoError(t, module.Start())
	assert.NoError(t, module.Stop())

	file, err := ioutil.TempFile("", "")

	assert.NoError(t, err)

	assert.NoError(t, file.Close())
	assert.NoError(t, os.Remove(file.Name()))

	module = newModule(file.Name(), transport, logrus.StandardLogger())

	assert.NotNil(t, module.tripperware)

	assert.Equal(t, "cache ("+file.Name()+")", module.Description())
	assert.NoError(t, module.Start())
	assert.NoError(t, module.Stop())

	assert.NoError(t, os.Remove(file.Name()))
}

func TestModule_oher(t *testing.T) {
	t.Parallel()

	transport := newTransport(t)
	module := newModule("", transport, logrus.StandardLogger())

	assert.NotPanics(t, func() { module.AddMetricSamples(nil) })
}

func TestModule_RoundTrip(t *testing.T) {
	t.Parallel()

	transport := newTransport(t)
	module := newModule("foo", transport, logrus.StandardLogger())

	req := new(http.Request)

	loc, _ := url.Parse("https://example.com")

	req.URL = loc

	res, err := module.RoundTrip(req) // nolint:bodyclose

	assert.NoError(t, err)
	assert.Same(t, req, res.Request)
}
