package plf

import "strings"

func Contains(substr string, vs ...string) []string {
	rst := make([]string, 0, len(vs))
	for _, v := range vs {
		if !strings.Contains(v, substr) {
			continue
		}

		rst = append(rst, v)
	}

	return rst
}
