// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package cache

import (
	"github.com/szkiba/xk6-cache/cache"
	"go.k6.io/k6/output"
)

func register() {
	output.RegisterExtension("cache", cache.New)
}

func init() { //nolint:gochecknoinits
	register()
}
