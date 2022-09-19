package plf

func FuncMap() map[string]any {
	return map[string]any{
		"regex":        Regex,
		"toSemver":     ToSemver,
		"semverLatest": SemverLatest,
		"semverMajorN": SemverMajorN,
		"semverMinorN": SemverMinorN,
	}
}
