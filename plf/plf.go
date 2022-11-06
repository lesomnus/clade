package plf

import (
	"reflect"

	"github.com/lesomnus/clade/sv"
	"github.com/lesomnus/pl"
)

func Funcs() pl.FuncMap {
	return pl.FuncMap{
		"semver":          Semver,
		"semverFinalized": SemverFinalized,
		"semverLatest":    SemverLatest,
		"semverMajorN":    SemverMajorN,
		"semverMinorN":    SemverMinorN,
		"semverPatchN":    SemverPatchN,
		"semverN":         SemverN,
	}
}

var default_convs pl.ConvMap = pl.ConvMap{
	reflect.TypeOf(""): map[reflect.Type]func(v reflect.Value) (any, error){
		reflect.TypeOf(&sv.Version{}): func(v reflect.Value) (any, error) {
			s := v.String()

			var rst *sv.Version = nil
			sv, err := sv.Parse(s)
			if err == nil {
				rst = &sv
			}

			return rst, nil
		},
	},
}

func Convs() pl.ConvMap {
	rst := make(pl.ConvMap)
	rst.MergeWith(default_convs)

	return rst
}
