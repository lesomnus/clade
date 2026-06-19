package tag_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/lesomnus/clade/tag"
)

func selectTags(t *testing.T, params string, tags []string) []tag.Matched {
	t.Helper()
	s, err := tag.New("semver", []byte(params))
	if err != nil {
		t.Fatalf("new selector: %v", err)
	}
	matched, err := s.Select(tags)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	return matched
}

func tagsOf(matched []tag.Matched) []string {
	out := make([]string, len(matched))
	for i, m := range matched {
		out[i] = m.Tag
	}
	return out
}

func TestSemverLastMajorMinor(t *testing.T) {
	tags := []string{"1.19.0", "1.20.1", "1.20.5", "1.21.0", "1.21.3", "2.0.0", "2.1.0", "latest"}
	matched := selectTags(t, "kind: semver\nlast-major: 2\nlast-minor: 2\n", tags)

	got := tagsOf(matched)
	want := []string{"2.1.0", "2.0.0", "1.21.3", "1.20.5"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("selected = %v, want %v", got, want)
	}
}

func TestSemverCollapsesToLatestPatch(t *testing.T) {
	tags := []string{"1.20.1", "1.20.5", "1.20.3"}
	matched := selectTags(t, "kind: semver\n", tags)
	if got := tagsOf(matched); len(got) != 1 || got[0] != "1.20.5" {
		t.Errorf("selected = %v, want [1.20.5]", got)
	}
}

func TestSemverPreReleaseSelectsExactMatch(t *testing.T) {
	tags := []string{
		"1.21.0", "1.21.0-alpine", "1.21.0-alpine3.20",
		"1.20.0-alpine", "1.22.0", "1.22.0-rc.1-alpine",
	}
	matched := selectTags(t, "kind: semver\npre-release: alpine\n", tags)

	got := tagsOf(matched)
	want := []string{"1.21.0-alpine", "1.20.0-alpine"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("selected = %v, want %v", got, want)
	}
}

func TestSemverDefaultExcludesPreRelease(t *testing.T) {
	tags := []string{"1.21.0", "1.21.0-alpine", "1.22.0-rc.1", "1.20.0-bookworm"}
	matched := selectTags(t, "kind: semver\n", tags)

	got := tagsOf(matched)
	want := []string{"1.21.0"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("selected = %v, want %v", got, want)
	}
}

func TestSemverDataRendersTemplate(t *testing.T) {
	matched := selectTags(t, "kind: semver\npre-release: alpine\n", []string{"1.22.3-alpine"})
	if len(matched) != 1 {
		t.Fatalf("got %d matches", len(matched))
	}

	if _, ok := matched[0].Data.(*semver.Version); !ok {
		t.Fatalf("Data is %T, want *semver.Version", matched[0].Data)
	}

	tmpl := template.Must(template.New("").Parse("{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"))
	var sb strings.Builder
	if err := tmpl.Execute(&sb, matched[0].Data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if sb.String() != "1.22.3-alpine" {
		t.Errorf("rendered = %q, want %q", sb.String(), "1.22.3-alpine")
	}
}
