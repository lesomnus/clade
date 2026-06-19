package version

type BuildInfo struct {
	Version  string
	GitRev   string
	GitDirty bool
}

//go:generate bash -c "../../scripts/gen-version-file.sh > /dev/null"
var build_info = BuildInfo{
	Version:  "YYMMDD-local",
	GitRev:   "0000000000000000000000000000000000000000",
	GitDirty: false,
}

func Get() BuildInfo {
	return build_info
}
