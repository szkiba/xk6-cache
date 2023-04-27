// SPDX-FileCopyrightText: 2021 - 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package cache

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

var instance *Module

const (
	moduleName = "cache"
	xk6Name    = "xk6-" + moduleName
)

func New(params output.Params) (output.Output, error) {
	return instance, nil
}

var envKey = "XK6_" + strings.ToUpper(moduleName)

func init() { //nolint:gochecknoinits
	file := os.Getenv(envKey)
	instance = newModule(file, http.DefaultTransport, logrus.StandardLogger())

	if file == "" {
		return
	}

	http.DefaultTransport = instance

	f, err := os.Open(file)
	if err == nil {
		if err := instance.tripperware.history.unmarshal(f); err != nil {
			panic(err)
		}
	}
}

type Module struct {
	logger      logrus.FieldLogger
	tripperware *tripperware
	recording   bool
	filename    string
}

func newModule(filename string, transport http.RoundTripper, logger logrus.FieldLogger) *Module {
	module := new(Module)

	module.logger = logger

	if filename == "" {
		return module
	}

	module.filename = filename
	module.tripperware = newTripperware(transport, logger)

	_, err := os.Stat(module.filename)

	module.recording = err != nil

	return module
}

func (m *Module) Description() string {
	return fmt.Sprintf("cache (%s)", m.filename)
}

func (m *Module) Start() error { return nil }

func (m *Module) Stop() error {
	if !m.recording {
		return nil
	}

	return m.tripperware.save(m.filename)
}

func (m *Module) AddMetricSamples(_ []metrics.SampleContainer) {}

func (m *Module) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.tripperware.RoundTrip(req)
}
