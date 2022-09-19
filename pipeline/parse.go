package pipeline

import (
	"errors"
	"fmt"
	"strings"
)

const symbols = "()|"

func readUntil(str string, r rune) (string, bool) {
	for i, c := range str {
		if c == r {
			return str[:i+1], true
		}
	}

	return "", false
}

func ReadToken(expr string) (int, string, error) {
	begin := -1
	for i, c := range expr {
		if c == ' ' {
			if begin < 0 {
				continue
			} else {
				return begin, expr[begin:i], nil
			}
		}
		if begin < 0 {
			begin = i

			if c == '"' {
				// Read quoted string.
				pos := 1 // Start after ".
				next := expr[begin+pos:]
				for {
					v, ok := readUntil(next, '"')
					if !ok {
						return -1, "", errors.New("unexpected end of expression: expected \"")
					}

					pos = pos + len(v)
					if len(v) > 1 && v[len(v)-2] == '\\' {
						// escaped double quotes.
						next = next[len(v):]
						continue
					}

					return begin, expr[begin : begin+pos], nil
				}
			}
		}

		if !strings.ContainsRune(symbols, c) {
			continue
		}

		if begin == i {
			return begin, expr[begin : begin+1], nil
		} else {
			return begin, expr[begin:i], nil
		}
	}

	if begin < 0 {
		return begin, "", nil
	} else {
		return begin, expr[begin:], nil
	}
}

func Parse(expr string) ([]*Fn, error) {
	scopes := []Pipeline{{nil}}

	pos := 0
	for {
		p, token, err := ReadToken(expr[pos:])
		if err != nil {
			return nil, fmt.Errorf("failed to read token at pos %d: %w", pos, err)
		}
		if p < 0 {
			break
		}

		pos += p

		scope := scopes[len(scopes)-1]
		cmd := scope[len(scope)-1]
		if err := func() error {
			if len(token) == 1 && strings.ContainsRune(symbols, rune(token[0])) {
				// symbols.
				if cmd == nil {
					return fmt.Errorf("expected a command name but it was symbol %c at pos %d", token[0], pos)
				}

				switch token[0] {
				case '(':
					scopes = append(scopes, Pipeline{nil})

				case ')':
					if len(scopes) < 2 {
						return fmt.Errorf("unexpected end of scope at pos %d", pos)
					}

					parent_scope := scopes[len(scopes)-2]
					parent_cmd := parent_scope[len(parent_scope)-1]
					parent_cmd.Args = append(parent_cmd.Args, scope)
					scopes = scopes[:len(scopes)-1]

				case '|':
					scope = append(scope, nil)
					scopes[len(scopes)-1] = scope
				}
			} else if cmd == nil {
				// command name.
				cmd = &Fn{Name: token, Args: make([]any, 0)}
				scope[len(scope)-1] = cmd
			} else {
				// command args.
				cmd.Args = append(cmd.Args, token)
			}

			return nil
		}(); err != nil {
			return nil, err
		}

		pos += len(token)
	}

	if len(scopes) != 1 {
		return nil, errors.New("scope does not closed")
	}

	rst := scopes[0]
	if rst[len(rst)-1] == nil {
		return nil, errors.New("expected a command name")
	}

	return rst, nil
}
