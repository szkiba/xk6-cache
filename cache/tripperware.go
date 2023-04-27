// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package cache

import (
	"bytes"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type tripperware struct {
	transport http.RoundTripper
	history   *history
	logger    logrus.FieldLogger
}

func newTripperware(transport http.RoundTripper, logger logrus.FieldLogger) *tripperware {
	return &tripperware{transport: transport, history: new(history), logger: logger}
}

func (tw *tripperware) shouldStore(res *http.Response) bool {
	if res.StatusCode != http.StatusOK {
		return false
	}

	mediatype, _, err := mime.ParseMediaType(res.Header.Get(hdrContentType))
	if err != nil {
		return true
	}

	return strings.HasPrefix(mediatype, "text") || strings.Contains(mediatype, "javascript")
}

func (tw *tripperware) RoundTrip(req *http.Request) (*http.Response, error) {
	log := tw.logger.WithField("url", req.URL.String())

	if rep, ok := tw.history.get(req.URL); ok {
		log.Debug("cache hit")

		return reply2response(req, rep), nil
	}

	log.Debug("cache miss")

	req.Header.Del(hdrAcceptEncoding) // avoid compressed response

	res, err := tw.transport.RoundTrip(req)
	if err != nil || !tw.shouldStore(res) {
		return res, err
	}

	rep, err := response2reply(res)
	if err != nil {
		return nil, err
	}

	loc := *req.URL

	addK6QueryParam(&loc)

	tw.history.put(&loc, rep)

	return res, nil
}

func (tw *tripperware) save(filename string) error {
	tw.logger.WithField("size", len(tw.history.store)).Debug("history summary")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err := tw.history.marshal(file); err != nil {
		return err
	}

	return file.Close()
}

func response2reply(resp *http.Response) (*reply, error) {
	body, err := duplicateBody(resp)
	if err != nil {
		return nil, err
	}

	return &reply{header: filterHeader(resp.Header), body: body}, nil
}

func reply2response(req *http.Request, rep *reply) *http.Response {
	return &http.Response{ // nolint:exhaustruct
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBuffer(rep.body)),
		ContentLength: int64(len(rep.body)),
		Request:       req,
		Header:        cloneHeader(rep.header),
	}
}

const hdrAcceptEncoding = "Accept-Encoding"
