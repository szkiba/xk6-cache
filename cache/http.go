// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package cache

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func cloneHeader(from http.Header) http.Header {
	header := make(http.Header)

	for key, value := range from {
		dup := make([]string, len(value))

		copy(dup, value)

		header[key] = dup
	}

	return header
}

func filterHeader(from http.Header) http.Header {
	header := make(http.Header)

	for key, value := range from {
		if !strings.HasPrefix(key, "Content") && !strings.HasPrefix(key, "Access-Control") && key != "Set-Cookie" {
			continue
		}

		dup := make([]string, len(value))

		copy(dup, value)

		header[key] = dup
	}

	return header
}

func duplicateBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dup := make([]byte, len(body))

	copy(dup, body)

	resp.Body = io.NopCloser(bytes.NewReader(dup))

	return body, err
}

func addK6QueryParam(loc *url.URL) {
	if loc.Query().Has(k6QueryVar) {
		return
	}

	query := k6QuerySuffix

	if loc.RawQuery != "" {
		query = loc.RawQuery + "&" + query
	}

	loc.RawQuery = query
}

const (
	k6QueryVar    = "_k6"
	k6QuerySuffix = k6QueryVar + "=1"
)
