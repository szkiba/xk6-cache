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
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	flatbuffers "github.com/google/flatbuffers/go"
)

type recorder struct {
	file    string
	entries map[string][]byte
}

func newRecorder(file string) *recorder {
	return &recorder{file: file, entries: make(map[string][]byte)}
}

func (r *recorder) put(req *http.Request, resp *http.Response) *http.Response {
	if !strings.Contains(req.URL.RawQuery, k6QueryParam) {
		query := k6QueryParam

		if req.URL.RawQuery != "" {
			query = req.URL.RawQuery + "&" + k6QueryParam
		}

		req.URL.RawQuery = query
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return newResponse(req, http.StatusInternalServerError, []byte{})
	}

	r.entries[req.URL.String()] = body

	return newResponse(req, http.StatusOK, body)
}

func (r *recorder) save() error {
	builder := flatbuffers.NewBuilder(0)
	keys := make([]string, 0, len(r.entries))

	for k := range r.entries {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	offsets := make([]flatbuffers.UOffsetT, 0, len(r.entries))

	for _, k := range keys {
		url := builder.CreateString(k)
		body := builder.CreateByteVector(r.entries[k])
		EntryStart(builder)
		EntryAddUrl(builder, url)
		EntryAddBody(builder, body)
		offsets = append(offsets, EntryEnd(builder))
	}

	CacheStartEntriesVector(builder, len(r.entries))

	for _, offset := range offsets {
		builder.PrependUOffsetT(offset)
	}

	entries := builder.EndVector(len(r.entries))

	CacheStart(builder)
	CacheAddEntries(builder, entries)

	builder.Finish(CacheEnd(builder))

	return ioutil.WriteFile(r.file, builder.FinishedBytes(), 0644) // nolint
}
