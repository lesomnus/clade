package compare

// DefaultFor returns the comparator chain to use for a source kind when a port
// declares no compare list:
//
//	container: created, falling back to digest.
//	http:      none (existence-only); a pinned version's artifact never changes,
//	           so an existing primary tag is up to date and a new version is
//	           caught by the missing-target check before any comparator runs.
//
// An unknown kind yields no default; such a port must declare its own chain.
func DefaultFor(sourceKind string) []Spec {
	switch sourceKind {
	case "http":
		return nil
	case "container", "":
		return []Spec{{Kind: "created"}, {Kind: "digest"}}
	default:
		return nil
	}
}
