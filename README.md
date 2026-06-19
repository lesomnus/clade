# clade

`clade` keeps your derived container images fresh with their upstreams.

You describe a set of images to build ("ports") that are based on upstream
images you care about. `clade` watches the upstream tags, figures out which of
your images are out of date, and rebuilds them on top of the latest upstream —
in dependency order, since one of your images can itself be the base of another.

## How it works

- Each **port** is a directory with a `Dockerfile`, build context, and a
  `port.yaml` that declares the upstream to track and how the built image is
  named.
- `clade outdated` lists upstream tags, resolves the corresponding target
  images, and emits a serializable **graph** of the targets that need building.
- `clade build` walks that graph in topological order, building each Dockerfile
  with the resolved upstream passed as the `BASE` build argument, then pushes.
- Your own images can be the base of other ports, so an upstream update cascades
  to every descendant.

Registry metadata is cached (a metadata lookup costs rate limit) and can be
inspected or cleared with `clade cache`; tag selection, "is it outdated?", and
the build backend are all pluggable.

## Install

```sh
go install github.com/lesomnus/clade@latest
# or build from a checkout
go build -o clade .
```

`clade build` shells out to `docker buildx`, so Docker with Buildx is required to
build (not to just compute the graph).

## Quick start

Create a port under `ports/`:

```
ports/
  golang-dev/
    Dockerfile
    port.yaml
```

```yaml
# ports/golang-dev/port.yaml
parent:
  repo: docker.io/library/golang
  target:
    kind: semver
    last-major: 1
    last-minor: 2
    pre-release: alpine
build:
  kind: build
  repo: ghcr.io/me/golang-dev
  tags:
    - "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
    - "{{.Major}}.{{.Minor}}-alpine"
```

```dockerfile
# ports/golang-dev/Dockerfile
ARG BASE
FROM ${BASE}
RUN go version
```

Then:

```sh
clade outdated                 # show which targets are stale
clade outdated --format json   # the build graph, serialized
clade build                    # build & push the stale targets, in order
clade build --dry-run          # print the buildx commands instead
```

A runnable example lives in [ports/golang-dev](ports/golang-dev).

## Automation

The workflows in [.github/workflows](.github/workflows) run `clade` on a
schedule: `refresh.yaml` computes the graph and dispatches a `build.yaml` run
per stale target, in topological order.

## Documentation

- [docs/architecture.md](docs/architecture.md) — packages and data flow
- [docs/cli.md](docs/cli.md) — commands and configuration
- [docs/port.md](docs/port.md) — the `port.yaml` reference

The build graph schema is documented inline in
[proto/clade/v1/graph.proto](proto/clade/v1/graph.proto).
