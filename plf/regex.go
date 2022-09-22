package plf

import (
	"fmt"
	"reflect"
	"regexp"
)

type RegexMatch map[string]any

func (r RegexMatch) String() string {
	v := reflect.ValueOf(r["$"])
	if v.Kind() != reflect.String {
		return ""
	} else {
		return v.String()
	}
}

func Regex(expr string, ss ...string) ([]RegexMatch, error) {
	pattern, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	rst := make([]RegexMatch, 0, len(ss))
	for _, s := range ss {
		matched := pattern.FindStringSubmatch(s)
		if matched == nil {
			continue
		}

		match := make(RegexMatch)
		match["$"] = s
		match["_"] = matched
		for i, name := range pattern.SubexpNames()[1:] {
			if name == "" {
				continue
			}

			match[name] = matched[i+1]
		}

		rst = append(rst, match)
	}

	return rst, nil
}
