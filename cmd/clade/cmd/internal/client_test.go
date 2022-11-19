package internal_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestRepositoryTagsAll(t *testing.T) {
	require := require.New(t)

	reg := newRegistry(t)
	reg.repos["repo/name"] = &repository{
		name: "repo/name",
		manifests: map[string]manifest{
			"a": {},
			"b": {},
			"c": {},
		},
	}

	s := httptest.NewTLSServer(reg.handler())
	defer s.Close()

	u, err := url.Parse(s.URL)
	require.NoError(err)

	named, err := reference.ParseNamed(fmt.Sprintf("%s/repo/name", u.Host))
	require.NoError(err)

	repo, err := internal.NewRepository(named, internal.WithTransport(s.Client().Transport))
	require.NoError(err)

	ctx := context.Background()
	tags, err := repo.Tags(ctx).All(ctx)
	require.NoError(err)

	require.ElementsMatch(reg.repos["repo/name"].Tags(), tags)
}

func TestManifestGetter(t *testing.T) {
	require := require.New(t)

	reg := newRegistry(t)
	reg.repos["repo/name"] = &repository{
		name: "repo/name",
		manifests: map[string]manifest{
			"tag": {
				contentType: "application/vnd.docker.distribution.manifest.list.v2+json",
				blob: []byte(`{
					"schemaVersion": 2,
					"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
					"manifests": [
						{
							"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
							"size": 7143,
							"digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
							"platform": {
								"architecture": "ppc64le",
								"os": "linux"
							}
						},
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
				contentType: "application/vnd.docker.distribution.manifest.v2+json",
				blob: []byte(`{
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

	s := httptest.NewTLSServer(reg.handler())
	defer s.Close()

	u, err := url.Parse(s.URL)
	require.NoError(err)

	named, err := reference.ParseNamed(fmt.Sprintf("%s/repo/name", u.Host))
	require.NoError(err)

	tagged, err := reference.WithTag(named, "tag")
	require.NoError(err)

	t.Run("returns manifest getter if tag is exists", func(t *testing.T) {
		ctx := context.Background()
		getter, err := internal.NewManifestGetter(ctx, tagged, internal.WithTransport(s.Client().Transport))
		require.NoError(err)

		t.Run("gets (fat) manifest", func(t *testing.T) {
			manifest, err := getter.Get(ctx)
			require.NoError(err)

			_, ok := manifest.(*manifestlist.DeserializedManifestList)
			require.True(ok)
		})

		t.Run("gets manifest with digest", func(t *testing.T) {
			manifest, err := getter.GetByDigest(ctx, digest.NewDigestFromEncoded(digest.SHA256, "5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"))
			require.NoError(err)

			_, ok := manifest.(*schema2.DeserializedManifest)
			require.True(ok)
		})
	})

	t.Run("returns ErrManifestUnknown if tag is not exists", func(t *testing.T) {
		tagged, err := reference.WithTag(named, "not_exists")
		require.NoError(err)

		ctx := context.Background()
		_, err = internal.NewManifestGetter(ctx, tagged, internal.WithTransport(s.Client().Transport))
		require.ErrorIs(err, internal.ErrManifestUnknown)
	})
}
