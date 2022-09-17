package pipeline

import (
	"fmt"
	"reflect"
	"strconv"

	"golang.org/x/exp/slices"
)

type Cmd struct {
	Name string
	Args []any
}

type Pipeline []*Cmd

type FuncMap map[string]any

type Executor struct {
	Funcs FuncMap
}

func implicitConvert(out reflect.Type, in reflect.Type, v reflect.Value) (any, error) {
	make_err := func() error {
		return fmt.Errorf("invalid conversion from %s to %s", in.Kind().String(), out.Kind().String())
	}

	switch in.Kind() {
	case reflect.String:
		switch out.Kind() {
		case reflect.Int:
			return strconv.Atoi(v.String())
		default:
			return nil, make_err()
		}
	default:
		return nil, make_err()
	}
}

func (e *Executor) invoke(fn any, args []any) (any, error) {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()

	// Check if number of argument is fit.
	num_fixed_args := ft.NumIn()
	if ft.IsVariadic() {
		num_fixed_args--
		if len(args) < num_fixed_args {
			return nil, fmt.Errorf("expected at least %d args but %d args are given", ft.NumIn()-1, len(args))
		}
	} else if len(args) != num_fixed_args {
		return nil, fmt.Errorf("expected %d args but %d args are given", ft.NumIn(), len(args))
	}

	input_args := make([]reflect.Value, len(args))
	for i, arg := range args {
		if pl, ok := arg.(Pipeline); ok {
			rst, err := e.Execute(pl)
			if err != nil {
				return nil, err
			}

			arg = rst
		}

		j := i
		if i >= num_fixed_args {
			j = num_fixed_args
		}

		t_arg := reflect.TypeOf(arg)
		t_in := ft.In(j)
		if i >= num_fixed_args {
			t_in = t_in.Elem()
		}
		if t_arg.AssignableTo(t_in) {
			input_args[i] = reflect.ValueOf(arg)
			continue
		}

		v, err := implicitConvert(t_in, t_arg, reflect.ValueOf(arg))
		if err != nil {
			return nil, err
		}

		input_args[i] = reflect.ValueOf(v)
	}

	rst := fv.Call(input_args)
	if len(rst) > 2 {
		return nil, fmt.Errorf("command have to return one or two values but %d values are returned", len(rst))
	} else if len(rst) == 2 {
		err, ok := rst[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("type of second return value of command must be an error but it was %s", rst[1].Type().Name())
		}
		return rst[0].Interface(), err
	} else {
		return rst[0].Interface(), nil
	}
}

func (e *Executor) Execute(pl Pipeline) (any, error) {
	prev := []any{}
	singular := true
	for _, cmd := range pl {
		fn, ok := e.Funcs[cmd.Name]
		if !ok {
			return nil, fmt.Errorf("unknown command %s", cmd.Name)
		}

		args := make([]any, len(cmd.Args)+len(prev))
		copy(args, cmd.Args)
		copy(args[len(cmd.Args):], prev)

		rst, err := e.invoke(fn, append(slices.Clone(cmd.Args), prev...))
		if err != nil {
			return nil, fmt.Errorf("failed invoke command %s: %w", cmd.Name, err)
		}

		singular = reflect.TypeOf(rst).Kind() != reflect.Slice
		if singular {
			prev = []any{rst}
		} else {
			v := reflect.ValueOf(rst)
			prev = make([]any, v.Len())
			for i := 0; i < v.Len(); i++ {
				prev[i] = v.Index(i).Interface()
			}
		}
	}

	if singular {
		return prev[0], nil
	} else {
		return prev, nil
	}
}
