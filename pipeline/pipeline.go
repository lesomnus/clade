package pipeline

import (
	"fmt"
	"reflect"
	"strconv"

	"golang.org/x/exp/slices"
)

type Fn struct {
	Name string
	Args []any
}

type Pipeline []*Fn

func (p Pipeline) hasFunction(fn *Fn, name string) bool {
	if fn.Name == name {
		return true
	}

	for _, arg := range fn.Args {
		if fn, ok := arg.(*Fn); !ok {
			continue
		} else if p.hasFunction(fn, name) {
			return true
		}
	}

	return false
}

func (p Pipeline) HasFunction(name string) bool {
	for _, fn := range p {
		if p.hasFunction(fn, name) {
			return true
		}
	}

	return false
}

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
		}
	default:
		switch out.Kind() {
		case reflect.String:
			if m, ok := in.MethodByName("String"); !ok {

			} else if t := m.Func.Type(); t.NumIn() == 1 && t.NumOut() == 1 && t.Out(0).Kind() == reflect.String {
				return m.Func.Call([]reflect.Value{v})[0].String(), nil
			}
		}
	}

	return nil, make_err()
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
		if rst[1].IsNil() {
			return rst[0].Interface(), nil
		}

		err, ok := rst[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("type of second return value of command must be an error but it was %s", rst[1].Type().Name())
		}
		return rst[0].Interface(), err
	} else {
		return rst[0].Interface(), nil
	}
}

func pass(args ...any) []any {
	return args
}

func Return(v any) Pipeline {
	return Pipeline{&Fn{Name: ">", Args: []any{v}}}
}

func (e *Executor) Execute(pl Pipeline) (any, error) {
	e.Funcs[">"] = pass

	prev := []any{}
	singular := true
	for _, fn := range pl {
		f, ok := e.Funcs[fn.Name]
		if !ok {
			return nil, fmt.Errorf("unknown command %s", fn.Name)
		}

		args := make([]any, len(fn.Args)+len(prev))
		copy(args, fn.Args)
		copy(args[len(fn.Args):], prev)

		rst, err := e.invoke(f, append(slices.Clone(fn.Args), prev...))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn.Name, err)
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
