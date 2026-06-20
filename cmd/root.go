package cmd

import (
	"context"

	"github.com/lesomnus/otx"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/flg"
	"github.com/lesomnus/xli/frm"
)

func NewCmdRoot() *xli.Command {
	return &xli.Command{
		Name: "clade",

		Flags: flg.Flags{
			&flg.String{Name: "config", Brief: "path to config file"},
		},

		Commands: []*xli.Command{
			NewCmdVersion(),
			NewCmdConfig(),
			NewCmdOutdated(),
			NewCmdGraph(),
			NewCmdBuild(),
			NewCmdCache(),
		},

		Handler: xli.Chain(
			xli.RequireSubcommand(),
			xli.OnRunPass(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
				if frm.HasSeq(frm.From(ctx).Next(), "version") {
					return next(ctx)
				}

				ctx, _, err := UseConfigInit(ctx, cmd)
				if err != nil {
					return err
				}

				o := otx.From(ctx)
				defer o.Shutdown(ctx)

				return next(ctx)
			}),
		),
	}
}
