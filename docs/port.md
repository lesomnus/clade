# Port

*Port* file is a file in YAML format that describes what the image's tags are and which images the tags depend on.

## Full Example of Port

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
      tag: ( remoteTags | semverFinalized | semverN 1 2 1 )

    # args:

    # dockerfile:
    # context:

  - tags:
      - ( printf "%d.%d-%s" $.Major $.Minor $.Pre[0] )
      - ( printf "%d-%s"    $.Major         $.Pre[0] )
    from:
      name: registry.hub.docker.com/library/golang
      tag: ( remoteTags | regex .+alpine$ | semverN 1 2 1 )

    args:
      USERNAME: somnus

    dockerfile: ./alpine/Dockerfile  
    context: ./alpine
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
      tag: ( remoteTags | semverFinalized | semverN 1 2 1 )

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
      tag: ( remoteTags | regex ".+alpine$" | semverN 1 2 1 )

    args:
      USERNAME: somnus
      UID: 1000

    dockerfile: /foo/bar/alpine/Dockerfile  
    context: /foo/bar/alpine
```


## Properties

Relative paths are resolved from the path of the *Port* file.

### Top-Level Object

| Property     | Type              | Description                                                                                                                                                      |
| ------------ | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`       | string            | **Required.** Name of the image this *Port* file creates.                                                                                                        |
| `args`       | {string: scalar}  | Variables passed when building the image. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details                   |
| `dockerfile` | filename          | Dockerfile to use for the build. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details. (Default: "./Dockerfile") |
| `context`    | dirname           | Docker build context. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details. (Default: ".")                       |
| `images`     | [Image](#image)[] | Description of the image reference this image depends on and the corresponding tag.                                                                              |

- `images[].args` is merged with `args`.
- If `images[].dockerfile` is empty, it will be set by `dockerfile`.
- If `images[].context` is empty, it will be set by `context`.

### Image

| Property     | Type                   | Description                                                                                                                                    |
| ------------ | ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `tags`       | (string \| Pipeline)[] | Tags of the image to be created. The resolved value by `from` is passed as data if it is a *Pipeline*.                                         |
| `from`       | string                 | Reference to the image this image depends on. Pipeline expression is allowed in the tag part.                                                  |
| `args`       | {string: scalar}       | Variables passed when building the image. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details |
| `dockerfile` | filename               | Dockerfile to use for the build. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details.         |
| `context`    | dirname                | Docker build context. See [this](https://docs.docker.com/engine/reference/commandline/build/#description) for more details.                    |
