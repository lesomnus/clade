package tag_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/lesomnus/clade/tag"
)

func calverSelect(t *testing.T, params string, tags []string) []tag.Matched {
	t.Helper()
	s, err := tag.New("calver", []byte(params))
	if err != nil {
		t.Fatalf("new selector: %v", err)
	}
	matched, err := s.Select(tags)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	return matched
}

func TestCalverLtsOnly(t *testing.T) {
	// Ubuntu LTS is released every even year in April.
	tags := []string{"20.04", "21.04", "22.04", "23.04", "23.10", "24.04", "24.10", "latest"}
	params := "kind: calver\nlayout: \"YY.0M\"\nwhere:\n  year: { mod: [2, 0] }\n  month: { in: [4] }\n"
	matched := calverSelect(t, params, tags)

	got := tagsOf(matched)
	want := []string{"24.04", "22.04", "20.04"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("selected = %v, want %v", got, want)
	}
}

func TestCalverLastKeepsNewest(t *testing.T) {
	tags := []string{"20.04", "22.04", "24.04"}
	params := "kind: calver\nlayout: \"YY.0M\"\nlast: 2\n"
	matched := calverSelect(t, params, tags)

	got := tagsOf(matched)
	want := []string{"24.04", "22.04"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("selected = %v, want %v", got, want)
	}
}

func TestCalverIgnoresNonMatching(t *testing.T) {
	tags := []string{"24.04", "24.04.1", "noble", "24-04", "v24.04"}
	params := "kind: calver\nlayout: \"YY.0M\"\n"
	matched := calverSelect(t, params, tags)

	if got := tagsOf(matched); len(got) != 1 || got[0] != "24.04" {
		t.Errorf("selected = %v, want [24.04]", got)
	}
}

func TestCalverPreservesZeroPadding(t *testing.T) {
	matched := calverSelect(t, "kind: calver\nlayout: \"YY.0M\"\n", []string{"24.04"})
	if len(matched) != 1 {
		t.Fatalf("got %d matches", len(matched))
	}

	cv, ok := matched[0].Data.(*tag.CalVer)
	if !ok {
		t.Fatalf("Data is %T, want *tag.CalVer", matched[0].Data)
	}
	if cv.Month != "04" {
		t.Errorf("Month = %q, want %q", cv.Month, "04")
	}

	tmpl := template.Must(template.New("").Parse("{{.Year}}.{{.Month}}"))
	var sb strings.Builder
	if err := tmpl.Execute(&sb, matched[0].Data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if sb.String() != "24.04" {
		t.Errorf("rendered = %q, want %q", sb.String(), "24.04")
	}
}

func TestCalverFullYearAndDay(t *testing.T) {
	tags := []string{"2024.01.09", "2023.12.31", "2024.1.9"}
	matched := calverSelect(t, "kind: calver\nlayout: \"YYYY.0M.0D\"\n", tags)

	got := tagsOf(matched)
	want := []string{"2024.01.09", "2023.12.31"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("selected = %v, want %v", got, want)
	}
}
