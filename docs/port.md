# Port

*Port* file is a file in YAML format that describes what the image's tags are and which images the tags depend on.

```yaml
name: ghcr.io/lesomnus/dev-golang
args:
  USERNAME: hypnos
  UID: 1000

dockerfile: ./Dockerfile
context: ./context
  
images:
  - tags:
      - ( printf "%d.%d" $.Major $.Minor )
      - ( printf "%d"    $.Major         )
    from:
      name: registry.hub.docker.com/library/golang
      tags: ( remoteTags | semverFinalized | semverN 1 2 1 )
      with:
        - name: registry.hub.docker.com/library/ubuntu
          tag: ( remoteTags | semverLatest )

    # args:

    # dockerfile:
    # context:
    # platform:

  - tags:
      - ( printf "%d.%d-%s" $.Major $.Minor $.Pre[0] )
      - ( printf "%d-%s"    $.Major         $.Pre[0] )
    from:
      name: registry.hub.docker.com/library/golang
      tags: ( remoteTags | regex .+alpine$ | semverN 1 2 1 )
      with:
        - name: registry.hub.docker.com/library/ubuntu
          tag: ( remoteTags | semverLatest )

    args:
      USERNAME: somnus

    dockerfile: ./alpine/Dockerfile  
    context: ./alpine
    platform: linux & amd64
```

If the path of the above *Port* file is `/foo/bar/port.yaml`, it is evaluated as follows:

```yaml
name: ghcr.io/lesomnus/dev-golang
args:
  USERNAME: hypnos
  UID: 1000

dockerfile: /foo/bar/Dockerfile
context: /foo/bar/context
  
images:
  - tags:
      - ( printf "%d.%d" $.Major $.Minor )
      - ( printf "%d"    $.Major         )
    from:
      name: registry.hub.docker.com/library/golang
      tags: ( remoteTags | semverFinalized | semverN 1 2 1 )

    args:
      USERNAME: hypnos
      UID: 1000

    dockerfile: /foo/bar/Dockerfile
    context: /foo/bar/context

  - tags:
      - ( printf "%d.%d-%s" $.Major $.Minor $.Pre[0] )
      - ( printf "%d-%s"    $.Major         $.Pre[0] )
    from:
      name: registry.hub.docker.com/library/golang
      tags: ( remoteTags | regex ".+alpine$" | semverN 1 2 1 )

    args:
      USERNAME: somnus
      UID: 1000

    dockerfile: /foo/bar/alpine/Dockerfile  
    context: /foo/bar/alpine
```


## Properties

Relative paths are resolved from the path of the *Port* file.

### Top-Level Object

| Property     | Type                | Description                                                                                                                                                      |
| ------------ | ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`       | `string`            | **Required.** Name of the image this *Port* file creates.                                                                                                        |
| `args`       | `map(string)`       | Variables passed when building the image. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details                   |
| `dockerfile` | `string`            | Dockerfile to use for the build. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details. (Default: "./Dockerfile") |
| `context`    | `string`            | Path to build context. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details. (Default: ".")                      |
| `platform`   | `string`            | Boolean expression indicating on which platform it can be built.                                                                                                 |
| `images`     | [`Image[]`](#image) | **Required.** Descriptions of the image to build.                                                                                                                |

- `images[].args` is merged with `args`.
- If `images[].dockerfile` is empty, it will be set by `dockerfile`.
- If `images[].context` is empty, it will be set by `context`.
- If `images[].platform` is empty, it will be set by `platform`.

### Image

| Property     | Type                               | Description                                                                                                                                    |
| ------------ | ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `tags`       | `string[]`                         | **Required.** Tags of the image to be created. The resolved value by `from` is passed as data if it is a *Pipeline*.                           |
| `from`       | `string` \| [`object`](#imagefrom) | **Required.** Canonical reference of the base image.                                                                                           |
| `args`       | `map(string)`                      | Variables passed when building the image. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details |
| `dockerfile` | `string`                           | Dockerfile to use for the build. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details.         |
| `context`    | `string`                           | Docker build context. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details.                    |
| `platform`   | `string`                           | Boolean expression indicating on which platform it can be built.                                                                               |

### Image.from

| Property | Type                                              | Description                                               |
| -------- | ------------------------------------------------- | --------------------------------------------------------- |
| `name`   | `string`                                          | **Required.** Name of the base image.                     |
| `tags`   | `string`                                          | **Required.** Tag of the base image. Pipeline is allowed. |
| `with`   | `string` \| [`ImageReference[]`](#imagereference) | Canonical reference of additional base images.            |

### ImageReference

| Property | Type     | Description                                               |
| -------- | -------- | --------------------------------------------------------- |
| `name`   | `string` | **Required.** Name of the base image.                     |
| `tag`    | `string` | **Required.** Tag of the base image. Pipeline is allowed. |

## Multiple Dependency

*CLade* provides the means to track multiple base images for a single image.
See [ports/dev-docker/Dockerfile](/ports/dev-docker/Dockerfile) which builds `lesomnus/dev-docker` image.
The image is based on `libarary/debian` image, and copies Docker engine binary from `library/docker` image.
This is useful because `library/docker` does not provide a Debian base image.
This image tracks tags for `library/docker`, but is also considered outdated when `library/debian` is updated.


### Usage

The following is part of [ports/dev-docker/port.yaml](/ports/dev-docker/port.yaml):

```yaml
...
images:
  - tags: ...
    from:
	    name: registry.hub.docker.com/library/docker
      tags: ( tags | semverFinalized | semverN 2 2 0 )
      with:
        - name: registry.hub.docker.com/library/debian
          tag: ( tags | semver | semverLatest )
...
```

The field `image[].from.with` is a list of additional base image references.
Note that `.with[].tag` must be evaluated as single value of tag to track.
Resolved image reference can be referenced in a builder-supplied environment variable with the last part of the name in all caps.
For example, ` registry.hub.docker.com/library/debian@sha256:...` can be accessed by `${DEBIAN}` in Dockerfile.
You can manually specify the environment variable name using the 'as' field.
You can use `${Foo}` in the following case:
```yaml
images:
  - ...
    from:
      ...
      with:
        - name: registry.hub.docker.com/library/debian
          tag: ( tags | semver | semverLatest )
		      as: Foo
...
```
