package pipeline

import (
	"errors"
	"fmt"
	"io"
)

func Parse(expr io.Reader) (Pipeline, error) {
	scopes := []Pipeline{{nil}}

	l := NewLexer(expr)
	if err := func() error {
		for {
			t, err := l.Lex()
			if err != nil {
				return err
			}
			if t.Token == TokenEOF {
				break
			}

			scope := scopes[len(scopes)-1]
			cmd := scope[len(scope)-1]

			if cmd == nil {
				if t.Token != TokenText {
					return errors.New("expected a function name")
				}

				cmd = &Fn{Name: t.Value, Args: []any{}}
				scope[len(scope)-1] = cmd
				continue
			}

			switch t.Token {

			// Beginning of scope.
			case TokenLeftParen:
				scopes = append(scopes, Pipeline{nil})

			// End of scope.
			case TokenRightParen:
				if len(scopes) < 2 {
					return errors.New("unexpected end of scope")
				}

				parent_scope := scopes[len(scopes)-2]
				parent_cmd := parent_scope[len(parent_scope)-1]
				parent_cmd.Args = append(parent_cmd.Args, scope)
				scopes = scopes[:len(scopes)-1]

			case TokenPipe:
				scopes[len(scopes)-1] = append(scope, nil)

			case TokenText:
				cmd.Args = append(cmd.Args, t.Value)

			case TokenString:
				cmd.Args = append(cmd.Args, t.Value[1:len(t.Value)-2])

			default:
				return fmt.Errorf("unknown token: %v", t)
			}
		}

		return nil
	}(); err != nil {
		return nil, fmt.Errorf("%d:%d: %w", l.pos.Line, l.pos.Column, err)
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
