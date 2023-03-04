package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/ocischema"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/graph"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

type OutdatedFlags struct {
	*RootFlags
	All bool
}

func CreateOutdatedCmd(flags *OutdatedFlags, svc Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "List outdated images",

		RunE: func(cmd *cobra.Command, args []string) error {
			bg := clade.NewBuildGraph()
			if err := svc.LoadBuildGraphFromFs(cmd.Context(), bg, flags.PortsPath); err != nil {
				return fmt.Errorf("load ports: %w", err)
			}

			var visit func(int, *graph.Node[*clade.ResolvedImage]) error
			visit = func(level int, node *graph.Node[*clade.ResolvedImage]) error {
				effective_next := make([]*graph.Node[*clade.ResolvedImage], 0, len(node.Next))
				for _, next := range node.Next {
					if next.Value.Skip {
						continue
					}

					if node.Key() != next.Value.From.Primary.String() {
						// Node is not primary dependency.
						continue
					}

					effective_next = append(effective_next, next)
				}

				visit_next := func() error {
					for _, node := range effective_next {
						if err := visit(level+1, node); err != nil {
							return fmt.Errorf("visit %s: %w", node.Key(), err)
						}
					}

					return nil
				}

				if len(node.Prev) == 0 {
					// level == 0
					return visit_next()
				}

				tagged, err := node.Value.Tagged()
				if err != nil {
					return err
				}
				if node.Key() != tagged.String() {
					// This node does not have a tag that is first one.
					return visit_next()
				}

				is_outdated, err := isOutdated(cmd.Context(), svc.Registry(), node)
				if err != nil {
					return fmt.Errorf(`check if "%s" is outdated: %w`, node.Key(), err)
				}
				if is_outdated {
					fmt.Fprintln(svc.Output(), node.Key())
					return nil
				}

				return visit_next()
			}

			for _, node := range bg.Roots() {
				visit(0, node)
			}

			return nil
		},
	}

	cmd_flags := cmd.Flags()
	cmd_flags.BoolVar(&flags.All, "all", false, "Print all images including skipped images")

	return cmd
}

var (
	outdated_flags = OutdatedFlags{RootFlags: &root_flags}
	outdated_cmd   = CreateOutdatedCmd(&outdated_flags, DefaultCmdService)
)

func init() {
	root_cmd.AddCommand(outdated_cmd)
}

func getManifest(ctx context.Context, reg Namespace, named reference.Named, dgst digest.Digest, opts ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	repo, err := reg.Repository(named)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}

	svc, err := repo.Manifests(ctx)
	if err != nil {
		return nil, fmt.Errorf("get manifest service: %w", err)
	}

	return svc.Get(ctx, dgst, opts...)
}

func getLayers(ctx context.Context, reg Namespace, ref reference.NamedTagged, dgst digest.Digest) ([]distribution.Descriptor, error) {
	opts := []distribution.ManifestServiceOption{}
	if dgst == digest.Digest("") {
		opts = append(opts, distribution.WithTag(ref.Tag()))
	}

	manif, err := getManifest(ctx, reg, ref, dgst, opts...)
	if err != nil {
		return nil, fmt.Errorf("get manifest: %w", err)
	}

	switch m := manif.(type) {
	case *ocischema.DeserializedManifest:
		return m.Layers, nil

	case *schema2.DeserializedManifest:
		return m.Layers, nil

	case *manifestlist.DeserializedManifestList:
		if len(m.Manifests) == 0 {
			return nil, fmt.Errorf("manifest list is empty")
		}

		layers, err := getLayers(ctx, reg, ref, m.Manifests[0].Digest)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", m.Manifests[0].Digest.String(), err)
		}

		return layers, nil
	}

	return nil, fmt.Errorf("unknown manifest type")
}

func isOutdatedByLayers(ctx context.Context, reg Namespace, node *graph.Node[*clade.ResolvedImage]) (bool, error) {
	ref, err := node.Value.Tagged()
	if err != nil {
		return false, err
	}

	layers, err := getLayers(ctx, reg, ref, digest.Digest(""))
	if err != nil {
		return false, fmt.Errorf("get layers: %w", err)
	}

	base_layers, err := getLayers(ctx, reg, node.Value.From.Primary, digest.Digest(""))
	if err != nil {
		return false, fmt.Errorf(`get layers of base image "%s": %w`, node.Value.From.Primary.String(), err)
	}

	if len(base_layers) > len(layers) {
		return true, nil
	}

	for i := 0; i < len(base_layers); i++ {
		if layers[i].Digest != base_layers[i].Digest {
			return true, nil
		}
	}

	return false, nil
}

func isOutdated(ctx context.Context, reg Namespace, node *graph.Node[*clade.ResolvedImage]) (bool, error) {
	tagged, err := node.Value.Tagged()
	if err != nil {
		return false, err
	}

	manif, err := getManifest(ctx, reg, tagged, digest.Digest(""), distribution.WithTag(tagged.Tag()))
	if err != nil {
		var errs errcode.Errors
		if errors.As(err, &errs) {
			if len(errs) == 0 {
				return false, errors.New("error is returned but the error is empty")
			}

			for _, err := range errs {
				// How to distinguish between no existence and permission denied?
				if errors.Is(err, v2.ErrorCodeManifestUnknown) ||
					errors.Is(err, v2.ErrorCodeNameUnknown) ||
					errors.Is(err, errcode.ErrorCodeDenied) {
					return true, nil
				}
			}
		}

		return false, fmt.Errorf("get manifest: %w", err)
	}

	deref_id := ""
	switch m := manif.(type) {
	case *ocischema.DeserializedManifest:
		if id, ok := m.Annotations[clade.AnnotationDerefId]; !ok {
			return isOutdatedByLayers(ctx, reg, node)
		} else {
			deref_id = id
		}

	case *schema2.DeserializedManifest:
		return isOutdatedByLayers(ctx, reg, node)

	case *manifestlist.DeserializedManifestList:
		if m.MediaType == manifestlist.SchemaVersion.MediaType {
			return isOutdatedByLayers(ctx, reg, node)
		}

		_, data, err := m.Payload()
		if err != nil {
			return false, fmt.Errorf("get manifestlist payload: %w", err)
		}

		var annotated struct {
			Annotations map[string]string `json:"annotations,omitempty"`
		}

		if err := json.Unmarshal(data, &annotated); err != nil {
			return false, fmt.Errorf("unmarshal oci index: %w", err)
		}

		if id, ok := annotated.Annotations[clade.AnnotationDerefId]; !ok {
			return isOutdatedByLayers(ctx, reg, node)
		} else {
			deref_id = id
		}

	default:
		return false, fmt.Errorf("unknown manifest type")
	}

	dgsts := make([][]byte, 0, 1+len(node.Value.From.Secondaries))
	for _, ref := range node.Value.From.All() {
		repo, err := reg.Repository(ref)
		if err != nil {
			return false, fmt.Errorf(`get repository of "%s": %w`, ref.String(), err)
		}

		desc, err := repo.Tags(ctx).Get(ctx, ref.Tag())
		if err != nil {
			return false, fmt.Errorf(`get digest of base image "%s": %w`, ref.String(), err)
		}

		dgst, err := hex.DecodeString(desc.Digest.Encoded())
		if err != nil {
			return false, fmt.Errorf(`parse digest of base image "%s": %s: %w`, ref.String(), desc.Digest.Encoded(), err)
		}

		dgsts = append(dgsts, dgst)
	}

	deref_id_curr := clade.CalcDerefId(dgsts...)
	return deref_id != deref_id_curr, nil
}
