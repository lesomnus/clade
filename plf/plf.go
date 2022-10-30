package plf

import "github.com/lesomnus/pl"

func Funcs() pl.FuncMap {
	return pl.FuncMap{
		"contains":     Contains,
		"regex":        Regex,
		"toSemver":     ToSemver,
		"semverLatest": SemverLatest,
		"semverMajorN": SemverMajorN,
		"semverMinorN": SemverMinorN,
		"semverPatchN": SemverPatchN,
		"semverN":      SemverN,
	}
}
