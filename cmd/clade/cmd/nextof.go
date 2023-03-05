package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

type NextofFlags struct {
	*RootFlags
}

func CreateNextofCmd(flags *NextofFlags, svc Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nextof [flags] plan reference",
		Short: "Make build plan",
		Args:  cobra.ExactArgs(2),

		RunE: func(cmd *cobra.Command, args []string) error {
			var ref reference.NamedTagged
			if named, err := reference.ParseNamed(args[1]); err != nil {
				return fmt.Errorf(`parse "%s": %w`, args[1], err)
			} else if tagged, ok := named.(reference.NamedTagged); !ok {
				return fmt.Errorf("reference must e tagged")
			} else {
				ref = tagged
			}

			var plan clade.BuildPlan
			if plan_str, err := os.ReadFile(args[0]); err != nil {
				return fmt.Errorf(`read plan file at "%s": %w`, args[0], err)
			} else if err := json.Unmarshal(plan_str, &plan); err != nil {
				return fmt.Errorf(`unmarshal plan file: %w`, err)
			}

			bg := clade.NewBuildGraph()
			if err := svc.LoadBuildGraphFromFs(cmd.Context(), bg, flags.PortsPath); err != nil {
				return fmt.Errorf("load ports: %w", err)
			}

			iter := -1
			var coll_curr []string
			for i, collections := range plan.Iterations {
				for _, collection := range collections {
					if !slices.Contains(collection, ref.String()) {
						continue
					}

					iter = i
					coll_curr = collection
					break
				}

				if coll_curr != nil {
					break
				}
			}

			if coll_curr == nil {
				return fmt.Errorf(`reference "%s" not found`, ref.String())
			}

			l := log.Ctx(cmd.Context())
			l.Info().Int("n", iter).Msg("current iteration")

			iter_next := iter + 1
			if len(plan.Iterations) <= iter_next {
				fmt.Fprintln(svc.Output(), "[]")
				return nil
			}

			var colls_next = make([][]string, 0)
			for _, collection := range plan.Iterations[iter_next] {
				for _, ref := range collection {
					node, ok := bg.Get(ref)
					if !ok {
						return fmt.Errorf(`get "%s": not in graph`, ref)
					}

					is_next_coll := false
					for _, base_ref := range node.Value.From.All() {
						fmt.Printf("base_ref.String(): %v\n", base_ref.String())
						is_next_coll = slices.Contains(coll_curr, base_ref.String())
						if is_next_coll {
							break
						}
					}

					if is_next_coll {
						colls_next = append(colls_next, collection)
						break
					}
				}
			}

			rst, err := json.Marshal(colls_next)
			if err != nil {
				return fmt.Errorf("marshal next collections: %w", err)
			}

			fmt.Fprintln(svc.Output(), string(rst))
			return nil
		},
	}

	return cmd
}

var (
	nextof_flags = NextofFlags{RootFlags: &root_flags}
	nextof_cmd   = CreateNextofCmd(&nextof_flags, DefaultCmdService)
)

func init() {
	root_cmd.AddCommand(nextof_cmd)
}
