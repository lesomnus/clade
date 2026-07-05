package tag

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

func init() {
	Register("calver", newCalver)
}

// calverConfig is the target spec for the calver strategy.
//
//	select:
//	  kind: calver
//	  layout: "YY.0M"   # how each version is parsed (see below)
//	  where:            # keep only versions whose components satisfy these
//	    year:  { mod: [2, 0] } # year % 2 == 0
//	    month: { in: [4] }     # month is April
//	  last: 2           # keep the newest 2 versions (0 = all)
//
// layout is a sequence of tokens separated by literal characters. A tag is
// matched only when it has the exact same literals in the same places.
// Recognized tokens:
//
//	YYYY  four-digit year   (2024)
//	YY    short year        (24)
//	0M    zero-padded month (04)
//	MM    month             (4)
//	0D    zero-padded day   (09)
//	DD    day               (9)
//	MICRO trailing number   (1)
//
// The captured text is preserved verbatim for rendering, so a zero-padded
// "04" stays "04" instead of collapsing to 4.
type calverConfig struct {
	Layout string `yaml:"layout"`
	Where  struct {
		Year  *calverPredicate `yaml:"year"`
		Month *calverPredicate `yaml:"month"`
		Day   *calverPredicate `yaml:"day"`
		Micro *calverPredicate `yaml:"micro"`
	} `yaml:"where"`
	Last int `yaml:"last"`
}

// calverPredicate constrains a single numeric component of a version. When both
// In and Mod are given, a value must satisfy both.
type calverPredicate struct {
	// In keeps only values contained in this set.
	In []int `yaml:"in"`
	// Mod is [divisor, remainder]; it keeps only values where value%divisor ==
	// remainder. Useful for cadences such as Ubuntu LTS (every even year).
	Mod []int `yaml:"mod"`
}

func (p *calverPredicate) allows(v int) bool {
	if p == nil {
		return true
	}
	if len(p.In) > 0 && !slices.Contains(p.In, v) {
		return false
	}
	if len(p.Mod) == 2 && v%p.Mod[0] != p.Mod[1] {
		return false
	}
	return true
}

// CalVer is a parsed calendar version. Its exported string fields hold the text
// exactly as it appeared in the tag (so "04" is preserved), while the build tag
// templates render from them, e.g. "{{.Year}}.{{.Month}}".
type CalVer struct {
	Version string // the whole tag, e.g. "24.04"
	Year    string // e.g. "24" or "2024"
	Month   string // e.g. "04" ("" when the layout has no month)
	Day     string // e.g. "09" ("" when the layout has no day)
	Micro   string // e.g. "1"  ("" when the layout has no micro)

	year, month, day, micro int
}

// less reports whether a sorts before b in ascending version order.
func (a *CalVer) less(b *CalVer) bool {
	if a.year != b.year {
		return a.year < b.year
	}
	if a.month != b.month {
		return a.month < b.month
	}
	if a.day != b.day {
		return a.day < b.day
	}
	return a.micro < b.micro
}

type calverToken struct {
	name string // "year", "month", "day" or "micro"
	// width is the exact digit count when strict; otherwise it is only the
	// nominal width used when the token is adjacent to another token. 0 means
	// unbounded (variable, never strict).
	width  int
	strict bool // consume exactly width digits, e.g. the zero-padded "04"
}

var calverTokens = map[string]calverToken{
	"YYYY":  {name: "year", width: 4, strict: true},
	"YY":    {name: "year", width: 2},
	"0Y":    {name: "year", width: 2, strict: true},
	"MM":    {name: "month", width: 2},
	"0M":    {name: "month", width: 2, strict: true},
	"DD":    {name: "day", width: 2},
	"0D":    {name: "day", width: 2, strict: true},
	"MICRO": {name: "micro", width: 0},
}

// calverSegment is either a literal separator or a token, never both.
type calverSegment struct {
	literal string
	token   *calverToken
}

type calverSelector struct {
	segments []calverSegment
	where    calverConfig
	last     int
}

func newCalver(params []byte) (Selector, error) {
	var cfg calverConfig
	if err := yaml.Unmarshal(params, &cfg); err != nil {
		return nil, fmt.Errorf("decode calver target: %w", err)
	}
	if cfg.Layout == "" {
		return nil, fmt.Errorf("calver target: layout is required")
	}

	segments, err := parseLayout(cfg.Layout)
	if err != nil {
		return nil, fmt.Errorf("calver target: %w", err)
	}
	for _, p := range []*calverPredicate{cfg.Where.Year, cfg.Where.Month, cfg.Where.Day, cfg.Where.Micro} {
		if p == nil || len(p.Mod) == 0 {
			continue
		}
		if len(p.Mod) != 2 {
			return nil, fmt.Errorf("calver target: mod must be [divisor, remainder]")
		}
		if p.Mod[0] == 0 {
			return nil, fmt.Errorf("calver target: mod divisor must not be zero")
		}
	}

	return &calverSelector{segments: segments, where: cfg, last: cfg.Last}, nil
}

// parseLayout splits a layout into literal separators and tokens. A variable
// width token (MICRO) must be the last token or be followed by a literal, so
// that the boundary is unambiguous.
func parseLayout(layout string) ([]calverSegment, error) {
	var segments []calverSegment
	var literal strings.Builder
	flush := func() {
		if literal.Len() > 0 {
			segments = append(segments, calverSegment{literal: literal.String()})
			literal.Reset()
		}
	}

	for i := 0; i < len(layout); {
		matched := ""
		for name := range calverTokens {
			if strings.HasPrefix(layout[i:], name) && len(name) > len(matched) {
				matched = name
			}
		}
		if matched == "" {
			literal.WriteByte(layout[i])
			i++
			continue
		}
		flush()
		tok := calverTokens[matched]
		segments = append(segments, calverSegment{token: &tok})
		i += len(matched)
	}
	flush()

	for i, seg := range segments {
		if seg.token != nil && seg.token.width == 0 {
			if i != len(segments)-1 && segments[i+1].token != nil {
				return nil, fmt.Errorf("variable-width token must be last or followed by a separator")
			}
		}
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("layout has no tokens")
	}
	return segments, nil
}

// parse matches tag against the layout, returning nil when it does not conform.
func (s *calverSelector) parse(tag string) *CalVer {
	cv := &CalVer{Version: tag}
	i := 0
	for si, seg := range s.segments {
		if seg.token == nil {
			if !strings.HasPrefix(tag[i:], seg.literal) {
				return nil
			}
			i += len(seg.literal)
			continue
		}

		var end int
		switch {
		case seg.token.strict: // padded token: consume exactly its width
			end = i + seg.token.width
			if end > len(tag) {
				return nil
			}
		case si+1 < len(s.segments) && s.segments[si+1].token == nil:
			rel := strings.Index(tag[i:], s.segments[si+1].literal)
			if rel < 0 {
				return nil
			}
			end = i + rel
		case si == len(s.segments)-1:
			end = len(tag)
		default: // adjacent to another token: consume a nominal width
			end = i + seg.token.width
			if end > len(tag) {
				return nil
			}
		}

		digits := tag[i:end]
		n, ok := atoiStrict(digits)
		if !ok {
			return nil
		}
		switch seg.token.name {
		case "year":
			cv.Year, cv.year = digits, n
		case "month":
			cv.Month, cv.month = digits, n
		case "day":
			cv.Day, cv.day = digits, n
		case "micro":
			cv.Micro, cv.micro = digits, n
		}
		i = end
	}
	if i != len(tag) {
		return nil // trailing characters
	}
	return cv
}

// atoiStrict parses s as a run of ASCII digits, rejecting signs and empties so
// that only genuine numeric components match.
func atoiStrict(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// Select parses each tag with the layout, drops those failing the where
// predicates, then keeps the newest last versions.
func (s *calverSelector) Select(tags []string) ([]Matched, error) {
	var versions []*CalVer
	for _, t := range tags {
		cv := s.parse(t)
		if cv == nil {
			continue // ignore tags that do not match the layout
		}
		if !s.where.Where.Year.allows(cv.year) ||
			!s.where.Where.Month.allows(cv.month) ||
			!s.where.Where.Day.allows(cv.day) ||
			!s.where.Where.Micro.allows(cv.micro) {
			continue
		}
		versions = append(versions, cv)
	}

	sort.Slice(versions, func(i, j int) bool { return versions[j].less(versions[i]) })
	if s.last > 0 && len(versions) > s.last {
		versions = versions[:s.last]
	}

	out := make([]Matched, 0, len(versions))
	for _, cv := range versions {
		out = append(out, Matched{Tag: cv.Version, Data: cv})
	}
	return out, nil
}
