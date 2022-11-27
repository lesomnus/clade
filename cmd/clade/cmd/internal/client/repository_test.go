package client_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestRepositoryTags(t *testing.T) {

	reg := registry.NewRegistry(t)

	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_rul, err := url.Parse(s.URL)
	require.NoError(t, err)

	named, err := reference.ParseNamed(reg_rul.Host + "/repo/name")
	require.NoError(t, err)

	name := reference.Path(named)

	reg.Repos[name] = &registry.Repository{
		Name: name,
		Manifests: map[string]registry.Manifest{
			"foo": {
				ContentType: "application/vnd.docker.distribution.manifest.list.v2+json",
				Blob: []byte(`{
					"schemaVersion": 2,
					"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
					"manifests": [
						{
							"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
							"size": 7682,
							"digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
							"platform": {
								"architecture": "amd64",
								"os": "linux",
								"features": [
									"sse4"
								]
							}
						}
					]
				}`),
			},
			"sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270": {
				ContentType: "application/vnd.docker.distribution.manifest.v2+json",
				Blob: []byte(`{
					"schemaVersion": 1,
					"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
					"name": "hello-world",
					"tag": "latest",
					"architecture": "amd64",
					"fsLayers": [
						{
							"blobSum": "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
						},
						{
							"blobSum": "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
						},
						{
							"blobSum": "sha256:cc8567d70002e957612902a8e985ea129d831ebe04057d88fb644857caa45d11"
						},
						{
							"blobSum": "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
						}
					],
					"history": [
						{
							"v1Compatibility": "{\"id\":\"e45a5af57b00862e5ef5782a9925979a02ba2b12dff832fd0991335f4a11e5c5\",\"parent\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"created\":\"2014-12-31T22:57:59.178729048Z\",\"container\":\"27b45f8fb11795b52e9605b686159729b0d9ca92f76d40fb4f05a62e19c46b4f\",\"container_config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [/hello]\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"docker_version\":\"1.4.1\",\"config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/hello\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":0}\n"
						},
						{
							"v1Compatibility": "{\"id\":\"e45a5af57b00862e5ef5782a9925979a02ba2b12dff832fd0991335f4a11e5c5\",\"parent\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"created\":\"2014-12-31T22:57:59.178729048Z\",\"container\":\"27b45f8fb11795b52e9605b686159729b0d9ca92f76d40fb4f05a62e19c46b4f\",\"container_config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [/hello]\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"docker_version\":\"1.4.1\",\"config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/hello\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":0}\n"
						}
					],
					"signatures": [
						{
							"header": {
								"jwk": {
									"crv": "P-256",
									"kid": "OD6I:6DRK:JXEJ:KBM4:255X:NSAA:MUSF:E4VM:ZI6W:CUN2:L4Z6:LSF4",
									"kty": "EC",
									"x": "3gAwX48IQ5oaYQAYSxor6rYYc_6yjuLCjtQ9LUakg4A",
									"y": "t72ge6kIA1XOjqjVoEOiPPAURltJFBMGDSQvEGVB010"
								},
								"alg": "ES256"
							},
							"signature": "XREm0L8WNn27Ga_iE_vRnTxVMhhYY0Zst_FfkKopg6gWSoTOZTuW4rK0fg_IqnKkEKlbD83tD46LKEGi5aIVFg",
							"protected": "eyJmb3JtYXRMZW5ndGgiOjY2MjgsImZvcm1hdFRhaWwiOiJDbjAiLCJ0aW1lIjoiMjAxNS0wNC0wOFQxODo1Mjo1OVoifQ"
						}
					]
				}`),
			},
		},
	}

	reg_client := client.NewDistRegistry()
	reg_client.Transport = s.Client().Transport
	reg_client.Cache = cache.NewMemCacheStore()

	repo, err := reg_client.Repository(named)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("tags are cached", func(t *testing.T) {
		require := require.New(t)

		defer reg_client.Cache.Clear()

		tags, err := repo.Tags(ctx).All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"foo"}, tags)

		cached_tags, ok := reg_client.Cache.GetTags(named)
		require.True(ok)
		require.ElementsMatch([]string{"foo"}, cached_tags)

		reg_client.Cache.SetTags(named, []string{"bar"})
		cached_tags, err = repo.Tags(ctx).All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"bar"}, cached_tags)
	})

	t.Run("get manifest", func(t *testing.T) {
		require := require.New(t)

		svc, err := repo.Manifests(ctx)
		require.NoError(err)

		manifest, err := svc.Get(ctx, digest.Digest(""), distribution.WithTag("foo"))
		require.NoError(err)

		manifest_list, ok := manifest.(*manifestlist.DeserializedManifestList)
		require.True(ok)

		manifests := manifest_list.References()
		require.Len(manifests, 1)
		require.Equal(digest.NewDigestFromEncoded(digest.SHA256, "5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"), manifests[0].Digest)

		manifest_child, err := svc.Get(ctx, manifests[0].Digest)
		require.NoError(err)

		_, ok = manifest_child.(*schema2.DeserializedManifest)
		require.True(ok)
	})

	t.Run("returns ErrorCodeManifestUnknown if tag is not exists", func(t *testing.T) {
		require := require.New(t)

		svc, err := repo.Manifests(ctx)
		require.NoError(err)

		_, err = svc.Get(ctx, digest.Digest(""), distribution.WithTag("not-exists"))
		require.Error(err)

		var errs errcode.Errors
		ok := errors.As(err, &errs)
		require.True(ok)
		require.Len(errs, 1)
		require.ErrorIs(errs[0], v2.ErrorCodeManifestUnknown)
	})
}
