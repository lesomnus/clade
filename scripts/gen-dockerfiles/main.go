// Command gen-dockerfiles renders each ports/<port>/Dockerfile.tmpl template
// into a sibling Dockerfile, expanding the shared partials in
// ports/_common/*.dockerfile.
//
// Each shared partial is a standalone *.dockerfile holding raw Dockerfile
// content (so editors highlight it natively); text/template names the partial
// after its filename, so a port references it as {{ template "apt.dockerfile" . }}.
//
// The generated Dockerfile files are committed; edit the Dockerfile.tmpl
// templates (and the shared partials) instead. Run via `go generate`.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

const (
	portsDir = "ports"
	// srcName is the template marker: ports/<port>/Dockerfile.tmpl renders to
	// ports/<port>/Dockerfile.
	srcName = "Dockerfile.tmpl"
	outName = "Dockerfile"
)

// blankRuns matches three or more consecutive newlines. Each partial file ends
// with a trailing newline, which—combined with the blank line a port leaves
// between {{ template }} calls—would yield a double blank line at every seam;
// collapsing keeps the output to a single separating blank line.
var blankRuns = regexp.MustCompile(`\n{3,}`)

// funcs are template helpers. dict builds a map from alternating key/value
// arguments so a Dockerfile can pass options to a shared partial, e.g.
// {{ template "apt.dockerfile" dict "SystemPython" false }}.
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
	partials, err := template.New("partials").Funcs(funcs).ParseGlob(filepath.Join(portsDir, "_common", "*.dockerfile"))
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
	// Collapse the double blank line left at each partial seam, then normalize
	// to a single trailing newline (the last partial contributes one of its own).
	rendered := blankRuns.ReplaceAll(buf.Bytes(), []byte("\n\n"))
	rendered = append(bytes.TrimRight(rendered, "\n"), '\n')
	return os.WriteFile(out, rendered, 0o644)
}
