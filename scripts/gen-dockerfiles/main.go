// Command gen-dockerfiles renders each ports/<port>/Dockerfile.tmpl template
// into a sibling Dockerfile, expanding the shared partials in
// ports/_common/*.tmpl.
//
// The generated Dockerfile files are committed; edit the Dockerfile.tmpl
// templates (and the shared partials) instead. Run via `go generate`.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const (
	portsDir = "ports"
	// srcName is the template marker: ports/<port>/Dockerfile.tmpl renders to
	// ports/<port>/Dockerfile.
	srcName = "Dockerfile.tmpl"
	outName = "Dockerfile"
)

// funcs are template helpers. dict builds a map from alternating key/value
// arguments so a Dockerfile$ can pass options to a shared partial, e.g.
// {{ template "apt" dict "SystemPython" false }}.
var funcs = template.FuncMap{
	"dict": func(pairs ...any) (map[string]any, error) {
		if len(pairs)%2 != 0 {
			return nil, fmt.Errorf("dict: want an even number of arguments, got %d", len(pairs))
		}
		m := make(map[string]any, len(pairs)/2)
		for i := 0; i < len(pairs); i += 2 {
			k, ok := pairs[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict: key %d is not a string", i)
			}
			m[k] = pairs[i+1]
		}
		return m, nil
	},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen-dockerfiles:", err)
		os.Exit(1)
	}
}

func run() error {
	partials, err := template.New("partials").Funcs(funcs).ParseGlob(filepath.Join(portsDir, "_common", "*.tmpl"))
	if err != nil {
		return fmt.Errorf("parse shared partials: %w", err)
	}

	srcs, err := filepath.Glob(filepath.Join(portsDir, "*", srcName))
	if err != nil {
		return err
	}
	for _, src := range srcs {
		if err := render(partials, src); err != nil {
			return fmt.Errorf("render %s: %w", src, err)
		}
		fmt.Println("generated", filepath.Join(filepath.Dir(src), outName))
	}
	return nil
}

func render(partials *template.Template, src string) error {
	body, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	t, err := partials.Clone()
	if err != nil {
		return err
	}
	if _, err := t.New(srcName).Parse(string(body)); err != nil {
		return err
	}

	var buf bytes.Buffer
	// The syntax parser directive must be the very first line of a Dockerfile,
	// so it precedes the generated-file notice.
	buf.WriteString("# syntax=docker/dockerfile:1\n")
	fmt.Fprintf(&buf, "# Code generated from %s by scripts/gen-dockerfiles; DO NOT EDIT.\n\n", srcName)
	if err := t.ExecuteTemplate(&buf, srcName, nil); err != nil {
		return err
	}

	out := filepath.Join(filepath.Dir(src), outName)
	return os.WriteFile(out, buf.Bytes(), 0o644)
}
