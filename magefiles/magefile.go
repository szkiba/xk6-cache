// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/princjef/mageutil/bintool"
	"github.com/princjef/mageutil/shellcmd"
)

var Default = All

var linter = bintool.Must(bintool.New(
	"golangci-lint{{.BinExt}}",
	"1.51.1",
	"https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.GOOS}}-{{.GOARCH}}{{.ArchiveExt}}",
))

func Lint() error {
	if err := linter.Ensure(); err != nil {
		return err
	}

	return linter.Command(`run`).Run()
}

func Test() error {
	return shellcmd.Command(`go test -count 1 -coverprofile=coverage.txt ./...`).Run()
}

func Build() error {
	return shellcmd.Command(`xk6 build --with github.com/szkiba/xk6-cache=.`).Run()
}

var k6 = sh.RunCmd("./k6", "run", "--quiet", "--no-summary", "--no-usage-report")

var reImport = regexp.MustCompile(`(import .* from ["'])(https://[^"']+)["'].*`)

func imports(filename string) ([]string, error) {
	all := []string{}

	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	subs := reImport.FindAllSubmatch(body, -1)

	for _, sub := range subs {
		if len(sub) < 3 {
			continue
		}

		all = append(all, string(sub[2]))
	}

	return all, nil
}

func checkImportIn(location string, filename string) error {
	body, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(fmt.Sprintf(`Content-Location: %s`, location))

	if !re.Match(body) {
		return errors.New(fmt.Sprintf("%s: missing %s", filename, location))
	}

	return nil
}

func runIt(filename string, outdir string) error {
	cache := filepath.Join(outdir, filepath.Base(filename)+".eml")

	if err := os.RemoveAll(cache); err != nil {
		return err
	}

	os.Setenv("XK6_CACHE", cache)
	os.RemoveAll(cache)

	if err := k6("--out", "cache", filename); err != nil {
		return err
	}

	err := k6("--out", "cache", filename)
	if err != nil {
		return err
	}

	urls, err := imports(filename)
	if err != nil {
		return err
	}

	if len(urls) == 0 {
		return errors.New("missing imports for file: " + filename)
	}

	for _, u := range urls {
		if err := checkImportIn(u, cache); err != nil {
			return err
		}
	}

	return nil
}

func It() error {
	if err := os.MkdirAll("build", 0o755); err != nil {
		return err
	}

	all, err := filepath.Glob("scripts/*.js")
	if err != nil {
		return err
	}

	for _, script := range all {
		if err := runIt(script, "build"); err != nil {
			return err
		}
	}

	return nil
}

func Coverage() error {
	return shellcmd.Command(`go tool cover -html=coverage.txt`).Run()
}

func glob(patterns ...string) (string, error) {
	buff := new(strings.Builder)

	for _, p := range patterns {
		m, err := filepath.Glob(p)
		if err != nil {
			return "", err
		}

		_, err = buff.WriteString(strings.Join(m, " ") + " ")
		if err != nil {
			return "", err
		}
	}

	return buff.String(), nil
}

func License() error {
	all, err := glob("*.go", "*/*.go", ".*.yml", ".gitignore", "*/.gitignore", ".github/workflows/*")
	if err != nil {
		return err
	}

	return shellcmd.Command(
		`reuse annotate --copyright "Iván Szkiba" --merge-copyrights --license MIT --skip-unrecognised ` + all,
	).Run()
}

func Clean() error {
	sh.Rm("magefiles/bin")
	sh.Rm("coverage.txt")
	sh.Rm("bin")
	sh.Rm("build")
	sh.Rm("vendor.eml")

	return nil
}

func All() error {
	if err := Lint(); err != nil {
		return err
	}

	if err := Test(); err != nil {
		return err
	}

	if err := Build(); err != nil {
		return err
	}

	return It()
}
