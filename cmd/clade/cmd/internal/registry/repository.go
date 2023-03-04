package registry

import (
	"context"
	"math/rand"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/ocischema"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/uuid"
	"github.com/opencontainers/go-digest"
)

type RepositoryStorage struct {
	Manifests map[string]distribution.Manifest
	Blobs     map[string]Blob
	Tags      map[string]distribution.Descriptor
}

type Repository struct {
	named   reference.Named
	Storage *RepositoryStorage
}

func NewRepository(named reference.Named) *Repository {
	return &Repository{
		named: named,
		Storage: &RepositoryStorage{
			Manifests: make(map[string]distribution.Manifest),
			Blobs:     make(map[string]Blob),
			Tags:      make(map[string]distribution.Descriptor),
		},
	}
}

func (r *Repository) PopulateLayer() distribution.Descriptor {
	desc := distribution.Descriptor{
		MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
	}

	data := make([]byte, rand.Uint64()%200+100)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	desc.Size = int64(len(data))
	desc.Digest = digest.FromBytes(data)

	return desc
}

func (r *Repository) addManifest(manif distribution.Manifest) (distribution.Descriptor, bool) {
	mt, data, err := manif.Payload()
	if err != nil {
		panic(err)
	}

	dgst := digest.FromBytes(data)
	if _, ok := r.Storage.Manifests[dgst.String()]; ok {
		return distribution.Descriptor{}, false
	}

	r.Storage.Manifests[dgst.String()] = manif

	desc := distribution.Descriptor{
		MediaType: mt,
		Size:      int64(len(data)),
		Digest:    dgst,
	}

	return desc, true
}

func (r *Repository) PopulateManifest() (distribution.Descriptor, *schema2.DeserializedManifest) {
	for {
		manif := schema2.Manifest{
			Versioned: schema2.SchemaVersion,
			Config:    distribution.Descriptor{},
		}

		num_layers := rand.Intn(5) + 3
		manif.Layers = make([]distribution.Descriptor, num_layers)
		for i := range manif.Layers {
			manif.Layers[i] = r.PopulateLayer()
		}

		m, err := schema2.FromStruct(manif)
		if err != nil {
			panic(err)
		}

		desc, ok := r.addManifest(m)
		if !ok {
			continue
		}

		return desc, m
	}
}

func (r *Repository) PopulateOciManifest() (distribution.Descriptor, *ocischema.DeserializedManifest) {
	for {
		manif := ocischema.Manifest{
			Versioned:   ocischema.SchemaVersion,
			Config:      distribution.Descriptor{},
			Annotations: make(map[string]string),
		}

		num_layers := rand.Intn(5) + 3
		manif.Layers = make([]distribution.Descriptor, num_layers)
		for i := range manif.Layers {
			manif.Layers[i] = r.PopulateLayer()
		}

		m, err := ocischema.FromStruct(manif)
		if err != nil {
			panic(err)
		}

		desc, ok := r.addManifest(m)
		if !ok {
			continue
		}

		return desc, m
	}
}

func (r *Repository) PopulateManifestList() (distribution.Descriptor, *manifestlist.DeserializedManifestList) {
	manifs := make([]manifestlist.ManifestDescriptor, 2)

	{
		desc, _ := r.PopulateManifest()
		manifs[0] = manifestlist.ManifestDescriptor{
			Descriptor: desc,
			Platform: manifestlist.PlatformSpec{
				Architecture: "amd64",
				OS:           "linux",
			},
		}
	}

	{
		desc, _ := r.PopulateManifest()
		manifs[1] = manifestlist.ManifestDescriptor{
			Descriptor: desc,
			Platform: manifestlist.PlatformSpec{
				Architecture: "arm64",
				OS:           "linux",
			},
		}
	}

	manif, err := manifestlist.FromDescriptors(manifs)
	if err != nil {
		panic(err)
	}

	_, data, err := manif.Payload()
	if err != nil {
		panic(err)
	}

	dgst := digest.FromBytes(data)
	r.Storage.Manifests[dgst.String()] = manif

	desc := distribution.Descriptor{
		MediaType: manif.MediaType,
		Size:      int64(len(data)),
		Digest:    dgst,
	}

	return desc, manif
}

func (r *Repository) PopulateImageWithTag(tag string) (reference.NamedTagged, distribution.Descriptor, distribution.Manifest) {
	tagged, err := reference.WithTag(r.named, tag)
	if err != nil {
		panic(err)
	}

	desc, manif := r.PopulateManifestList()
	r.Storage.Tags[tagged.Tag()] = desc
	return tagged, desc, manif
}

func (r *Repository) PopulateImage() (reference.NamedTagged, distribution.Descriptor, distribution.Manifest) {
	var tagged reference.NamedTagged
	for tagged == nil {
		tag := uuid.Generate().String()
		t, err := reference.WithTag(r.named, tag)
		if err != nil {
			continue
		}

		_, ok := r.Storage.Tags[tag]
		if ok {
			continue
		}

		tagged = t
		break
	}

	desc, manif := r.PopulateManifestList()
	r.Storage.Tags[tagged.Tag()] = desc
	return tagged, desc, manif
}

func (r *Repository) Named() reference.Named {
	return r.named
}

func (r *Repository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	return &ManifestService{Repo: r}, nil
}

func (r *Repository) Blobs(ctx context.Context) distribution.BlobStore {
	return &BlobStore{Repo: r}
}

func (r *Repository) Tags(ctx context.Context) distribution.TagService {
	return &TagService{Repo: r}
}
