// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package cache

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"sync"
)

type reply struct {
	header http.Header
	body   []byte
}

type history struct {
	store map[string]*reply
	mu    sync.RWMutex
}

func (c *history) put(key *url.URL, value *reply) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if value.header == nil {
		value.header = http.Header{}
	}

	str := key.String()

	value.header.Set(hdrContentLocation, str)
	value.header.Set(hdrContentLength, strconv.Itoa(len(value.body)))

	loc := *key

	loc.RawQuery = ""

	cd := mime.FormatMediaType("attachment", map[string]string{"filename": loc.String()})

	value.header.Set(hdrContentDisposition, cd)

	if c.store == nil {
		c.store = make(map[string]*reply)

		hdr := http.Header{}
		hdr.Set(hdrContentType, cacheBodyContentType)

		entry := &reply{
			header: hdr,
			body:   []byte(cacheBody),
		}

		c.store[""] = entry
	}

	c.store[str] = value
}

func (c *history) get(key *url.URL) (*reply, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.store == nil {
		return nil, false
	}

	ret, ok := c.store[key.String()]

	return ret, ok
}

func (c *history) marshalHeader(writer io.Writer) error {
	hdr := http.Header{}
	hdr.Set(hdrSubject, cacheSubject)
	hdr.Set(hdrContentType, mime.FormatMediaType("multipart/mixed", map[string]string{"boundary": cacheBoundary}))

	if err := hdr.Write(writer); err != nil {
		return err
	}

	_, err := writer.Write([]byte("\r\n"))

	return err
}

func (c *history) marshal(writer io.Writer) error {
	if err := c.marshalHeader(writer); err != nil {
		return err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := []string{}

	for key := range c.store {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	out := multipart.NewWriter(writer)

	out.SetBoundary(cacheBoundary) // nolint:errcheck

	for _, key := range keys {
		entry := c.store[key]

		part, err := out.CreatePart(textproto.MIMEHeader(entry.header))
		if err != nil {
			return err
		}

		if _, err := part.Write(entry.body); err != nil {
			return err
		}
	}

	return out.Close()
}

func (c *history) unmarshalHeader(reader io.Reader) (string, io.Reader, error) {
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		return "", nil, err
	}

	mediatype, params, err := mime.ParseMediaType(msg.Header.Get(hdrContentType))
	if err != nil {
		return "", nil, err
	}

	if mediatype != "multipart/mixed" {
		return "", nil, errInvalidCacheContentType
	}

	boundary, ok := params["boundary"]
	if !ok {
		return "", nil, fmt.Errorf("%w: missing boundary parameter", errInvalidCacheContentType)
	}

	return boundary, msg.Body, nil
}

func (c *history) unmarshal(reader io.Reader) error {
	boundary, body, err := c.unmarshalHeader(reader)
	if err != nil {
		return err
	}

	inp := multipart.NewReader(body, boundary)

	for {
		part, err := inp.NextRawPart()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		cl := part.Header.Get(hdrContentLength)
		if len(cl) == 0 {
			continue
		}

		rep := new(reply)

		rep.header = http.Header(part.Header)

		body, err := io.ReadAll(part)
		if err != nil {
			return err
		}

		rep.body = body

		key, err := url.Parse(rep.header.Get(hdrContentLocation))
		if err != nil {
			return err
		}

		c.put(key, rep)
	}

	return nil
}

const (
	cacheBoundary         = "______________________________o_o______________________________"
	hdrContentLocation    = "Content-Location"
	hdrContentDisposition = "Content-Disposition"
	hdrContentType        = "Content-Type"
	hdrContentLength      = "Content-Length"
	hdrSubject            = "Subject"
	cacheBodyContentType  = "text/plain; charset=utf-8"
	cacheSubject          = xk6Name
	cacheBody             = `This is ` + xk6Name + `'s standard email format cache file that can be viewed with an email client such as Mozilla Thunderbird. Modules stored as email attachments.`
)

var errInvalidCacheContentType = errors.New("invalid cache Content-Type")
