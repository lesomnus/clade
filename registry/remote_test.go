package registry_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	ggcrreg "github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	creg "github.com/lesomnus/clade/registry"
)

func TestRemoteAgainstInMemoryRegistry(t *testing.T) {
	srv := httptest.NewServer(ggcrreg.New())
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	img, err := random.Image(256, 1)
	if err != nil {
		t.Fatalf("random image: %v", err)
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		t.Fatalf("config file: %v", err)
	}
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	cfg.Created = v1.Time{Time: created}
	cfg.Config.Labels = map[string]string{"foo": "bar"}
	img, err = mutate.ConfigFile(img, cfg)
	if err != nil {
		t.Fatalf("mutate config: %v", err)
	}

	repo := host + "/team/app"
	ref := repo + ":1.0.0"
	reference, err := name.ParseReference(ref, name.Insecure)
	if err != nil {
		t.Fatalf("parse ref: %v", err)
	}
	if err := remote.Write(reference, img); err != nil {
		t.Fatalf("push: %v", err)
	}

	want_digest, err := img.Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}

	r := creg.NewRemote(creg.WithInsecure(true))
	ctx := context.Background()

	tags, err := r.Tags(ctx, repo)
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	if len(tags) != 1 || tags[0] != "1.0.0" {
		t.Errorf("tags = %v, want [1.0.0]", tags)
	}

	info, err := r.Stat(ctx, ref)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Digest != want_digest.String() {
		t.Errorf("digest = %q, want %q", info.Digest, want_digest.String())
	}
	if !info.Created.Equal(created) {
		t.Errorf("created = %v, want %v", info.Created, created)
	}
	if info.Labels["foo"] != "bar" {
		t.Errorf("labels = %v, want foo=bar", info.Labels)
	}

	if _, err := r.Stat(ctx, repo+":absent"); !errors.Is(err, creg.ErrNotExist) {
		t.Errorf("stat absent err = %v, want ErrNotExist", err)
	}
}
