package pipeline

import (
	"errors"

	"gopkg.in/yaml.v3"
)

func (f *Fn) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		var expr string
		if err := n.Decode(&expr); err != nil {
			return err
		}

		pl, err := Parse(expr)
		if err != nil {
			return err
		}

		if len(pl) == 1 {
			f.Name = pl[0].Name
			f.Args = pl[0].Args
		} else {
			f.Name = ">"
			f.Args = []any{pl}
		}

	case yaml.SequenceNode:
		var tokens []string
		if err := n.Decode(&tokens); err != nil {
			return err
		}

		if l := len(tokens); l == 0 {
			return errors.New("empty expression")
		} else if l == 1 {
			f.Name = tokens[0]
			f.Args = []any{}
		} else {
			f.Name = tokens[0]
			f.Args = make([]any, l-1)

			// TODO: support nested pipeline?
			for i, token := range tokens[1:] {
				f.Args[i] = token
			}
		}

	default:
		return errors.New("invalid node type")
	}

	return nil
}
