package plf

import (
	"fmt"
	"regexp"
)

func Regex(expr string, ss ...string) ([]map[string]any, error) {
	pattern, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	rst := make([]map[string]any, 0, len(ss))
	for _, s := range ss {
		matched := pattern.FindStringSubmatch(s)
		if matched == nil {
			continue
		}

		match := make(map[string]any)
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
