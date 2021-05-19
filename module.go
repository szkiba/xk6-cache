// MIT License
//
// Copyright (c) 2021 Iván Szkiba
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/output"
	"go.k6.io/k6/stats"
)

const (
	moduleName   = "cache"
	metricPrefix = "xk6_" + moduleName
)

var envKey = "XK6_" + strings.ToUpper(moduleName)

// Register the extensions on module initialization.
func init() {
	file := os.Getenv(envKey)
	mod := New(file, http.DefaultTransport, logrus.StandardLogger())

	modules.Register("k6/x/"+moduleName, mod)
	output.RegisterExtension(moduleName, func(params output.Params) (output.Output, error) {
		mod.log = params.Logger

		return mod, nil
	})

	if file != "" {
		http.DefaultTransport = mod
	}
}

type Module struct {
	log       logrus.FieldLogger
	cache     *Cache
	recorder  *recorder
	transport http.RoundTripper
	hit       uint32
	miss      uint32
	entries   uint32
}

func New(file string, transport http.RoundTripper, log logrus.FieldLogger) *Module {
	mod := new(Module)

	mod.log = log

	if file == "" {
		return mod
	}

	mod.transport = transport

	if _, err := os.Stat(file); err == nil {
		if mod.cache, err = newCache(file); err != nil {
			log.Error(err)
		} else {
			mod.entries = uint32(mod.cache.EntriesLength())
		}
	} else {
		mod.recorder = newRecorder(file)
	}

	return mod
}

func (m *Module) Description() string {
	if m.recorder == nil {
		return "cache (-)"
	}

	return fmt.Sprintf("cache (%s)", m.recorder.file)
}

func (m *Module) Start() (err error) {
	if m.recorder != nil {
		return m.recorder.save()
	}

	return nil
}

func (m *Module) Stop() error {
	return nil
}

func (m *Module) AddMetricSamples(_ []stats.SampleContainer) {
}

func (m *Module) Measure(ctx context.Context, prefix string) bool {
	state := lib.GetState(ctx)

	if prefix == "" {
		prefix = metricPrefix
	}

	return stats.PushIfNotDone(ctx, state.Samples, stats.Samples{
		newSample(prefix, "hit", m.hit),
		newSample(prefix, "miss", m.miss),
		newSample(prefix, "entry", m.entries),
	})
}

func newSample(prefix, name string, value uint32) stats.Sample {
	return stats.Sample{
		Metric: stats.New(fmt.Sprintf("%s_%s_count", prefix, name), stats.Counter),
		Tags:   nil,
		Time:   time.Now(),
		Value:  float64(value),
	}
}

func (m *Module) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.cache != nil {
		if resp := m.cache.get(req); resp != nil {
			atomic.AddUint32(&m.hit, 1)

			return resp, nil
		}

		atomic.AddUint32(&m.miss, 1)
	}

	resp, err := m.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == http.StatusOK && m.recorder != nil {
		return m.recorder.put(req, resp), nil
	}

	return resp, err
}

func newResponse(req *http.Request, status int, body []byte) *http.Response {
	return &http.Response{
		Status:        http.StatusText(status),
		StatusCode:    status,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        make(http.Header),
	}
}
