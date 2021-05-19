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
)

//go:generate flatc --gen-onefile --go --go-namespace cache  cache.fbs

func newCache(file string) (*Cache, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return GetRootAsCache(data, 0), nil
}

func (c *Cache) get(req *http.Request) *http.Response {
	url := req.URL.String()

	n := c.EntriesLength()
	idx := sort.Search(n, func(i int) bool {
		e := &Entry{}
		c.Entries(e, i)

		return strings.Compare(url, string(e.Url())) >= 0
	})

	if idx < n {
		e := &Entry{}
		c.Entries(e, idx)

		if url == string(e.Url()) {
			return newResponse(req, http.StatusOK, e.BodyBytes())
		}
	}

	return nil
}

const k6QueryParam = "_k6=1"
