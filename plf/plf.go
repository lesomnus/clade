package plf

func FuncMap() map[string]any {
	return map[string]any{
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
