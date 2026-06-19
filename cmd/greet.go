package cmd

import (
	"context"
	"fmt"

	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/arg"
	"github.com/lesomnus/xli/flg"
)

func NewCmdGreet() *xli.Command {
	return &xli.Command{
		Name:  "greet",
		Brief: "greet someone by name",

		Args: arg.Args{
			&arg.String{Name: "name", Brief: "name of the person to greet"},
		},
		Flags: flg.Flags{
			&flg.String{Name: "format", Brief: "greeting format string (%s is replaced with name)"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			flg.VisitP(cmd, "format", &c.Greet.Format)

			name := arg.MustGet[string](cmd, "name")
			cmd.Println(fmt.Sprintf(c.Greet.Format, name))
			return nil
		}),
	}
}
