# CLade

[![test](https://github.com/lesomnus/clade/actions/workflows/test.yaml/badge.svg)](https://github.com/lesomnus/clade/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lesomnus/clade)](https://goreportcard.com/report/github.com/lesomnus/clade)
[![codebeat badge](https://codebeat.co/badges/3f52fdb4-c46d-4e79-b2d5-11c9ee89e80d)](https://codebeat.co/projects/github-com-lesomnus-clade-main)
[![codecov](https://codecov.io/gh/lesomnus/clade/branch/main/graph/badge.svg?token=8CBVDL7AW7)](https://codecov.io/gh/lesomnus/clade)

Keep your container images up to date.

*CLade* allows you to manage multiple Dockerfiles as a dependency tree and list images older than the upstream image.

## Installation

```
$ go install github.com/lesomnus/clade/cmd/clade@latest
```

## Usage

Command `clade` reads *Port* files named `port.yaml` in the directories of `ports` directory to construct a dependency tree. *Port* file describes what an image's tag is and which image that tag depends on.
```sh
$ mkdir -p ports/my-gcc
$ code ports/my-gcc/port.yaml
```

```yaml
# ports/my-gcc/port.yaml
name: ghcr.io/my_name/my-gcc

images:
  - tags: ['my-tag']
    from: registry.hub.docker.com/library/gcc:latest
```

The above *Port* file describes that `ghcr.io/my_name/my-gcc:my-tag` is built from the `registry.hub.docker.com/library/gcc:latest`.
Let's see if this parses well:

```sh
$ clade tree
registry.hub.docker.com/library/gcc:latest
        ghcr.io/my_name/my-gcc:my-tag
```

Good. But this is the most basic usage of *CLade*.
How can we create a new image with that version name whenever a new version of GCC is updated? Probably the simplest way would be to populate a list of `images` for all versions. Alternatively, there is a way to have *CLade* fetch tags from the remote repository. Updates our *Port* file as:

```yaml
# ports/my-gcc/port.yaml
name: ghcr.io/my_name/my-gcc

images:
  - tags: ( printf "%d.%d" $.Major $.Minor )
    from:
      name: registry.hub.docker.com/library/gcc
      tag: ( tags | semverLatest )
```

```sh
$ clade tree
registry.hub.docker.com/library/gcc:12.2
        ghcr.io/my_name/my-gcc:12.2
```
What happened? Where did *12.2* come from? Let's find out one by one. First, what `( remoteTags | semverLatest )` is? This is [pipeline expression](https://github.com/lesomnus/pl). The result of the previous function becomes the argument of the next function. So it means, fetch the tags from `registry.hub.docker.com`, then take the latest semver. That would be *12.2* at this point. The result of the pipeline is *Semver* type and it is passed to pipeline in `tags` as a data and the result of the pipeline become result tags.

If there is more than one result of the `from` pipeline, *CLade* generates that many images. Let's create `my-gcc:12.X` for all gcc 12 versions using `semverMajorN` which filters last N major versions. Also if there are more than one *tag* or *tag* pipeline is provided, multiple tags will be created from the same image.


```yaml
# ports/my-gcc/port.yaml
name: ghcr.io/my_name/my-gcc

images:
  - tags:
      - ( printf "%d.%d" $.Major $.Minor )
      - ( printf "%d"    $.Major         )
    from:
      name: registry.hub.docker.com/library/gcc
      tag: ( tags | semverMajorN 1 )
```

```sh
$ clade tree
registry.hub.docker.com/library/gcc:12
        ghcr.io/my_name/my-gcc:12.0
registry.hub.docker.com/library/gcc:12.1
        ghcr.io/my_name/my-gcc:12.1
registry.hub.docker.com/library/gcc:12.2
        ghcr.io/my_name/my-gcc:12.2
        ghcr.io/my_name/my-gcc:12
```

You can now track upstream container images. To build, you can use the `clade build` command. Before that, let's create a *Dockerfile* first.

```sh
code ports/my-gcc/Dockerfile
```

```Dockerfile
# ports/my-gcc/Dockerfile
ARG TAG
ARG BASE
FROM ${BASE}:${TAG}

ARG USERNAME=my_name
ARG USER_UID=${UID:-1000}
ARG USER_GID=${GID:-${USER_UID}}
RUN groupadd ${USER_GID} --gid ${USER_GID} \
	&& useradd \
		--create-home ${USERNAME} \
		--shell /bin/bash \
		--uid ${USER_UID} \
		--gid ${USER_GID}

WORKDIR /home/${USERNAME}

USER ${USERNAME}
```

Note that arguments `TAG` and `BASE` are upstream container image name and tag.
The command for the default option, `clade build` simply spawns `docker` command with proper arguments.
Let's see what *CLade* runs:
```sh
$ clade build --dry-run ghcr.io/my_name/my-gcc:12
[/usr/bin/docker build --file /path/to/ports/my-gcc/port.yaml/Dockerfile --tag ghcr.io/my_name/my-gcc:12.2 --tag ghcr.io/my_name/my-gcc:12 --build-arg BASE=registry.hub.docker.com/library/gcc --build-arg TAG=12.2 /path/to/ports/my-gcc]
```

## Materials

- More examples of *Port* file → [ports](ports)
- Full *Port* file reference → [docs/port.md](docs/port.md)
- Available pipeline functions → [docs/pipeline-functions.md](docs/pipeline-functions.md)
