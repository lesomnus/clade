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
	fixed_args := args
	if ft.IsVariadic() {
		// TODO:
		panic("not supported")

		num_fixed_args := ft.NumIn() - 1
		if len(args) < num_fixed_args {
			return nil, fmt.Errorf("expected at least %d args but %d args are given", ft.NumIn()-1, len(args))
		}

		fixed_args = args[:num_fixed_args]
	} else if len(args) != ft.NumIn() {
		return nil, fmt.Errorf("expected %d args but %d args are given", ft.NumIn(), len(args))
	}

	input_args := make([]reflect.Value, len(fixed_args))
	for i, arg := range fixed_args {
		if _, ok := arg.(Pipeline); ok {
			// Execute nested pipeline later since it is expensive.
			continue
		}

		t_arg := reflect.TypeOf(arg)
		t_in := ft.In(i)
		if t_arg.Kind() == t_in.Kind() {
			input_args[i] = reflect.ValueOf(arg)
			continue
		}

		v, err := implicitConvert(t_in, t_arg, reflect.ValueOf(arg))
		if err != nil {
			return nil, err
		}

		input_args[i] = reflect.ValueOf(v)
	}

	// TODO: implement variadic

	// Execute nested pipeline.
	for i, arg := range fixed_args {
		pl, ok := arg.(Pipeline)
		if !ok {
			continue
		}

		rst, err := e.Execute(pl)
		if err != nil {
			return nil, err
		}

		input_args[i] = reflect.ValueOf(rst)
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
	var next any = nil
	for _, cmd := range pl {
		fn, ok := e.Funcs[cmd.Name]
		if !ok {
			return nil, fmt.Errorf("unknown command %s", cmd.Name)
		}

		args := slices.Clone(cmd.Args)
		if next != nil {
			args = append(args, next)
		}

		rst, err := e.invoke(fn, args)
		if err != nil {
			return nil, fmt.Errorf("failed invoke command %s: %w", cmd.Name, err)
		}

		next = rst
	}

	return next, nil
}
