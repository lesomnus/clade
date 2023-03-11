package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/graph"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type PlanFlags struct {
	*RootFlags
}

func CreatePlanCmd(flags *PlanFlags, svc Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [flags] [reference... | -]",
		Short: "Make build plan",

		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.Ctx(cmd.Context())

			bg := clade.NewBuildGraph()
			if err := svc.LoadBuildGraphFromFs(cmd.Context(), bg, flags.PortsPath); err != nil {
				return fmt.Errorf("load ports: %w", err)
			}

			bg_filtered := clade.NewBuildGraph()

			var refs []reference.NamedTagged
			if len(args) == 0 {
				// Plan for all nodes.
				bg_filtered = bg
			} else {
				ref_args := args
				if (len(args) == 1) && (args[0] == "-") {
					ref_args = make([]string, 0)
					scanner := bufio.NewScanner(svc.Input())
					scanner.Split(bufio.ScanWords)

					for scanner.Scan() {
						ref_arg := scanner.Text()
						l.Info().Str("reference", ref_arg).Msg("parse from input stream")
						ref_args = append(ref_args, ref_arg)
					}
				}

				refs = make([]reference.NamedTagged, 0, len(ref_args))
				for _, arg := range ref_args {
					named, err := reference.ParseNamed(arg)
					if err != nil {
						return fmt.Errorf(`"%s": %w`, arg, err)
					}

					tagged, ok := named.(reference.NamedTagged)
					if !ok {
						return fmt.Errorf(`"%s": reference must be tagged`, arg)
					}

					refs = append(refs, tagged)
				}
			}

			var visit func(node *graph.Node[*clade.ResolvedImage]) error
			visit = func(node *graph.Node[*clade.ResolvedImage]) error {
				if node_existing, ok := bg_filtered.Get(node.Key()); ok && node_existing.Value != nil {
					return nil
				}

				if _, err := bg_filtered.Put(node.Value); err != nil {
					return fmt.Errorf(`put "%s": %w`, node.Key(), err)
				}

				for _, tag := range node.Value.Tags {
					ref, err := reference.WithTag(node.Value.Named, tag)
					if err != nil {
						panic(fmt.Sprintf(`"%s" must have valid tag: %s`, node.Value.Name(), err.Error()))
					}

					sibling, ok := bg.Get(ref.String())
					if !ok {
						panic(fmt.Sprintf(`"%s", sibling of "%s", must be exist`, node.Key(), ref.String()))
					}

					for _, next := range sibling.Next {
						return visit(next)
					}
				}

				return nil
			}

			for _, ref := range refs {
				key := ref.String()
				node, ok := bg.Get(key)
				if !ok {
					return fmt.Errorf(`get "%s": not in graph`, key)
				}

				visit(node)
			}

			bp := clade.NewBuildPlan(bg_filtered)
			rst, err := json.Marshal(bp)
			if err != nil {
				return fmt.Errorf("marshal build plan: %w", err)
			}

			fmt.Fprintln(svc.Output(), string(rst))
			return nil
		},
	}

	return cmd
}

var (
	plan_flags = PlanFlags{RootFlags: &root_flags}
	plan_cmd   = CreatePlanCmd(&plan_flags, DefaultCmdService)
)

func init() {
	root_cmd.AddCommand(plan_cmd)
}
