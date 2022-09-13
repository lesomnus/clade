package clade

import (
	"errors"

	"github.com/blang/semver/v4"
)

func SemverStringLatest(vs []string) (string, error) {
	rst := semver.Version{}
	ok := false
	for _, v := range vs {
		sv, err := semver.ParseTolerant(v)
		if err != nil {
			continue
		}

		if rst.LT(sv) {
			rst = sv
			ok = true
		}
	}

	if ok {
		return rst.String(), nil
	} else {
		return "", errors.New("valid semver is not found")
	}
}
