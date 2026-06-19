# `port.yaml` reference

A **port** is one buildable image. It lives in its own directory together with
its `Dockerfile`, build context, and a `port.yaml`:

```
ports/
  golang-dev/
    Dockerfile
    port.yaml
    ...context files
```

By default `clade` scans the `ports/` directory; each immediate subdirectory
that contains a `port.yaml` is a port (others are ignored).

```yaml
parent:
  repo: docker.io/library/golang
  target:
    kind: semver
    last-major: 1
    last-minor: 2
    match: "-alpine$"
build:
  kind: build
  repo: ghcr.io/me/golang-dev
  tag: "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
  # ...optional build options below
```

Required fields: `parent.repo`, `parent.target.kind`, `build.repo`, `build.tag`.

## `parent`

The upstream image to track.

| Field | Description |
| --- | --- |
| `repo` | Upstream repository, e.g. `docker.io/library/golang`. May also be the `build.repo` of another port — see [Chaining](#chaining-ports). |
| `target` | How upstream tags are selected. `target.kind` picks the strategy; the remaining fields are strategy-specific. |

### Tag selection — `kind: semver`

`semver` is the built-in selection strategy.

| Field | Description |
| --- | --- |
| `last-major` | Keep this many of the newest major lines. `0` (default) keeps all. |
| `last-minor` | Keep this many of the newest minor lines within each kept major. `0` (default) keeps all. |
| `match` | A [RE2](https://github.com/google/re2/wiki/Syntax) regular expression. Only tags matching it are considered. |

Selection works as follows:

1. Filter tags by `match` (if set).
2. Parse each with semver; tags that do not parse are ignored. Partial versions
   are accepted (`1.22` is treated as `1.22.0`).
3. Collapse to the newest version per `(major, minor)` line.
4. Keep the newest `last-major` major lines, and within each, the newest
   `last-minor` minor lines.

So `last-major: 1, last-minor: 2, match: "-alpine$"` against a golang repo keeps
the two newest minor lines of the newest major, alpine variants only.

> **Note.** Tags are matched as plain strings, so `match: "-alpine$"` also
> matches pre-release tags such as `1.27-rc-alpine`, and non-semver schemes
> (e.g. date tags) can parse in surprising ways. Constrain `match` accordingly.

The selected version is exposed to the `build.tag` template (see below).

## `build`

How the produced image is named and built.

| Field | Description |
| --- | --- |
| `repo` | Destination repository to push to. |
| `tag` | A Go [text/template](https://pkg.go.dev/text/template) rendered once per selected upstream tag. |
| `kind` | Build strategy: `build` (default, `docker buildx build`) or `bake` (`docker buildx bake`). |

### `tag` template

The template is rendered with the data of each selected upstream tag. For the
`semver` strategy the data is the parsed version, so these are available:

| Expression | Example for `1.22.3-alpine` |
| --- | --- |
| `{{.Major}}` `{{.Minor}}` `{{.Patch}}` | `1` `22` `3` |
| `{{.Prerelease}}` | `alpine` |
| `{{.Metadata}}` | (build metadata after `+`, if any) |
| `{{.Original}}` | `1.22.3-alpine` |
| `{{.String}}` | `1.22.3-alpine` |

For example `tag: "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"` turns upstream
`1.22.3-alpine` into target `ghcr.io/me/golang-dev:1.22.3-alpine`.

### Build options

The fields below are optional and shared by both `build` and `bake` kinds (they
map to `docker buildx` options). Paths are relative to the port directory.

| Field | Maps to | Notes |
| --- | --- | --- |
| `dockerfile` | `-f` | Default `Dockerfile`. |
| `context` | build context | Default `.` (the port directory). |
| `target` | `--target` | Dockerfile stage. |
| `platforms` | `--platform` | e.g. `[linux/amd64, linux/arm64]`. |
| `args` | `--build-arg` | `BASE` is injected automatically. |
| `labels` | `--label` | Base name/digest labels are injected automatically. |
| `annotations` | `--annotation` | |
| `cache-from` | `--cache-from` | e.g. `[type=gha]`. |
| `cache-to` | `--cache-to` | e.g. `[type=gha,mode=max]`. |
| `secrets` | `--secret` | |
| `ssh` | `--ssh` | |
| `no-cache` | `--no-cache` | |
| `pull` | `--pull` | |
| `provenance` | `--provenance` | |
| `sbom` | `--sbom` | |
| `network` | `--network` | |
| `add-hosts` | `--add-host` | |
| `allow` | `--allow` | |
| `extra-args` | appended verbatim | Escape hatch for options not modeled above. |

## The `BASE` argument

`clade` injects the resolved upstream reference as the `BASE` build argument, so
the Dockerfile builds *on top of the tracked upstream*:

```dockerfile
ARG BASE
FROM ${BASE}
# e.g. BASE = docker.io/library/golang:1.22.3-alpine
RUN go version
```

Each built image is also labelled automatically:

- `org.opencontainers.image.base.name` — the upstream reference.
- `org.opencontainers.image.base.digest` — the upstream digest (used by the
  `digest` outdated strategy).

## Chaining ports

When a port's `parent.repo` equals the `build.repo` of another port, an
**internal edge** is created: the downstream port tracks the tags produced by the
upstream port instead of a registry, and `clade` builds them in order.

```yaml
# ports/base/port.yaml
build: { kind: build, repo: ghcr.io/me/base, tag: "{{.Major}}.{{.Minor}}" }
parent: { repo: docker.io/library/debian, target: { kind: semver, last-major: 1 } }
```

```yaml
# ports/app/port.yaml  — built on top of ghcr.io/me/base
build: { kind: build, repo: ghcr.io/me/app, tag: "{{.Major}}.{{.Minor}}" }
parent: { repo: ghcr.io/me/base, target: { kind: semver } }
```

If `ghcr.io/me/base` is rebuilt, every descendant (`ghcr.io/me/app`, ...) is
considered outdated and rebuilt on top of the fresh base.
