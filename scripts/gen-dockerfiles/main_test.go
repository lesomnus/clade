package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"text/template"
)

// renderApt renders the real ports/_common/apt.dockerfile partial, invoked with
// the given dict expression exactly as a port's Dockerfile.tmpl would — e.g.
// `dict "Exclude" (list "eza")`. Tests run with the working directory at the
// package dir, so the shared partials live two levels up.
func renderApt(t *testing.T, dictExpr string) string {
	t.Helper()

	partials, err := template.New("partials").Funcs(funcs).
		ParseGlob(filepath.Join("..", "..", portsDir, "_common", "*.dockerfile"))
	if err != nil {
		t.Fatalf("parse partials: %v", err)
	}
	driver, err := partials.New("driver").Parse(`{{ template "apt.dockerfile" ` + dictExpr + ` }}`)
	if err != nil {
		t.Fatalf("parse driver %q: %v", dictExpr, err)
	}

	var buf bytes.Buffer
	if err := driver.ExecuteTemplate(&buf, "driver", nil); err != nil {
		t.Fatalf("execute %q: %v", dictExpr, err)
	}
	return buf.String()
}

// aptPkgLine matches one rendered package line: two tabs, the name, " \".
var aptPkgLine = regexp.MustCompile(`(?m)^\t\t(\S+) \\$`)

func aptPackages(out string) []string {
	ms := aptPkgLine.FindAllStringSubmatch(out, -1)
	pkgs := make([]string, len(ms))
	for i, m := range ms {
		pkgs[i] = m[1]
	}
	return pkgs
}

func contains(pkgs []string, name string) bool {
	for _, p := range pkgs {
		if p == name {
			return true
		}
	}
	return false
}

// TestAptPackageList drives the real partial through the dict→Exclude/Include
// pipeline and asserts on the rendered package set: excludes drop, includes add,
// and the result stays sorted. It mirrors how every port imports the partial.
func TestAptPackageList(t *testing.T) {
	base := aptPackages(renderApt(t, `dict`))
	if len(base) == 0 {
		t.Fatal("base render produced no packages")
	}

	tcs := []struct {
		desc     string
		dictExpr string
		present  []string
		absent   []string
	}{
		{
			desc:     "nil data keeps the base set (golang via .)",
			dictExpr: `.`,
			present:  []string{"eza", "python3", "python3-pip"},
		},
		{
			desc:     "empty dict keeps the base set",
			dictExpr: `dict`,
			present:  []string{"eza", "python3", "python3-pip"},
		},
		{
			desc:     "exclude drops eza, keeps python (node)",
			dictExpr: `dict "Exclude" (list "eza")`,
			present:  []string{"python3", "python3-pip", "git"},
			absent:   []string{"eza"},
		},
		{
			desc:     "exclude drops the system python (python)",
			dictExpr: `dict "Exclude" (list "python3" "python3-pip")`,
			present:  []string{"eza"},
			absent:   []string{"python3", "python3-pip"},
		},
		{
			desc:     "include adds a package in sorted order",
			dictExpr: `dict "Include" (list "neovim")`,
			present:  []string{"neovim", "eza"},
		},
		{
			desc:     "exclude and include compose",
			dictExpr: `dict "Exclude" (list "eza") "Include" (list "neovim")`,
			present:  []string{"neovim", "git"},
			absent:   []string{"eza"},
		},
		{
			desc:     "excluding an absent package is a no-op",
			dictExpr: `dict "Exclude" (list "not-there")`,
			present:  base,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			pkgs := aptPackages(renderApt(t, tc.dictExpr))

			if !sort.StringsAreSorted(pkgs) {
				t.Errorf("packages not sorted: %v", pkgs)
			}
			for _, p := range tc.present {
				if !contains(pkgs, p) {
					t.Errorf("want %q present, got %v", p, pkgs)
				}
			}
			for _, p := range tc.absent {
				if contains(pkgs, p) {
					t.Errorf("want %q absent, got %v", p, pkgs)
				}
			}
		})
	}
}

// TestAptRenderIsWellFormed locks the rendered shape of the real partial: the
// {{- -}} trims strip the package-list setup, and the range must emit one
// backslash-continued line per package so the RUN stays a single command.
func TestAptRenderIsWellFormed(t *testing.T) {
	out := renderApt(t, `dict "Exclude" (list "eza")`)

	if !strings.HasPrefix(out, "RUN --mount=") {
		t.Fatalf("expected output to start with the RUN instruction, got:\n%s", out)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if last := lines[len(lines)-1]; last != "\t&& locale-gen" {
		t.Errorf("expected last line %q, got %q", "\t&& locale-gen", last)
	}
	// Every line but the last continues the shell command; a dropped backslash
	// would split the RUN into broken commands.
	for _, l := range lines[:len(lines)-1] {
		if !strings.HasSuffix(l, " \\") {
			t.Errorf("line missing trailing continuation: %q", l)
		}
		if strings.TrimSpace(l) == "" {
			t.Error("unexpected blank line in apt block")
		}
	}
}

// TestHelpersDoNotMutateInputs guards the list helpers against in-place edits:
// the base list is rebuilt per render, but a careless filter or sort would still
// be a latent trap for any future caller that reuses a slice.
func TestHelpersDoNotMutateInputs(t *testing.T) {
	without := funcs["without"].(func(s, drop []string) []string)
	sortStrings := funcs["sortStrings"].(func(s []string) []string)

	base := []string{"c", "a", "b"}
	snapshot := append([]string(nil), base...)

	without(base, []string{"b"})
	sortStrings(base)

	for i := range snapshot {
		if base[i] != snapshot[i] {
			t.Fatalf("input mutated at %d: got %q, want %q", i, base[i], snapshot[i])
		}
	}
}
