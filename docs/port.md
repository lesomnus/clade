# `port.yaml` reference

A **port** is one buildable image. It lives in its own directory together with
its `Dockerfile`, build context, and a `port.yaml`:

```
ports/
  dev-golang/
    Dockerfile
    port.yaml
    ...context files
```

By default `clade` scans the `ports/` directory; each immediate subdirectory
that contains a `port.yaml` is a port (others are ignored).

```yaml
name: dev-golang   # optional; defaults to the port's directory name
source:
  kind: container
  repo: docker.io/library/golang
select:
  kind: semver
  last-major: 1
  last-minor: 2
  pre-release: alpine
build:
  kind: build
  repo: ghcr.io/me/dev-golang
  tags:
    - "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
    - "{{.Major}}.{{.Minor}}-alpine"
  # ...optional build options below
```

Required fields: `source.kind`, `select.kind`, `build.repo`, `build.tags`.
A `container` source also requires `source.repo`; an `http` source requires
`source.url`.

## `name`

An optional display name for the port, shown by `clade outdated`. When omitted it
defaults to the port's directory name (e.g. `dev-golang` for `ports/dev-golang`).

## `source`

Where the upstream versions to track come from. `source.kind` picks the
discovery strategy; the remaining fields are strategy-specific.

### `kind: container`

Lists the tags of an OCI repository as candidate versions.

| Field | Description |
| --- | --- |
| `repo` | Upstream repository, e.g. `docker.io/library/golang`. May also be the `build.repo` of another port — see [Chaining](#chaining-ports). |

A container source also provides the **base image**: the selected tag is
injected as the `BASE` build argument (see [The `BASE` argument](#the-base-argument)).

### `kind: http`

Fetches a single version string from a URL. The response body is expected to be
a bare version, e.g. `1.2.3`.

| Field | Description |
| --- | --- |
| `url` | Endpoint returning the latest version, e.g. `https://downloads.claude.ai/claude-code-releases/stable`. |

An `http` source has **no base image**: `clade` injects no `BASE` build-arg, so
the Dockerfile declares its own `FROM`. Because there is no upstream image to
compare against, an http target is judged outdated purely by **existence** — see
[`compare`](#compare). For this to detect a new release, make the first
(primary) `build.tag` the full `{{.Major}}.{{.Minor}}.{{.Patch}}` so that a new
version produces a primary tag absent in the destination repository.

## `select`

How the discovered versions are selected. `select.kind` picks the strategy; the
remaining fields are strategy-specific.

### `kind: semver`

`semver` is the built-in selection strategy. It also **parses** each version so
the `build.tags` templates can use the version components — so it applies even
when a source yields a single version.

| Field | Description |
| --- | --- |
| `last-major` | Keep this many of the newest major lines. `0` (default) keeps all. |
| `last-minor` | Keep this many of the newest minor lines within each kept major. `0` (default) keeps all. |
| `pre-release` | Keep only tags whose semver [pre-release](https://semver.org/#spec-item-9) is *exactly* this. Empty (default) keeps only plain releases with no pre-release. |

Selection works as follows:

1. Parse each version with semver; values that do not parse are ignored. Partial
   versions are accepted (`1.22` is treated as `1.22.0`).
2. Keep versions whose pre-release exactly equals `pre-release` (empty by default).
3. Collapse to the newest version per `(major, minor)` line.
4. Keep the newest `last-major` major lines, and within each, the newest
   `last-minor` minor lines.

So `last-major: 1, last-minor: 2, pre-release: alpine` against a golang repo
keeps the two newest minor lines of the newest major, `-alpine` variants only.

> **Note.** The match is exact, so `pre-release: alpine` selects `1.22.3-alpine`
> but **not** `1.22.3-alpine3.20` (pre-release `alpine3.20`) nor
> `1.24.0-rc.3-alpine` (pre-release `rc.3-alpine`). Likewise the default empty
> value excludes every pre-release, e.g. `-rc.1`, `-bookworm`, `-windowsservercore-*`.

The selected version is exposed to the `build.tags` templates (see below).

## `build`

How the produced image is named and built.

| Field | Description |
| --- | --- |
| `repo` | Destination repository to push to. |
| `tags` | A list of Go [text/templates](https://pkg.go.dev/text/template), each rendered once per selected version. The built image is tagged with every rendered tag. |
| `kind` | Build strategy: `build` (default, `docker buildx build`) or `bake` (`docker buildx bake`). |

### `tags` templates

Each template is rendered with the data of each selected version. For the
`semver` strategy the data is the parsed version, so these are available:

| Expression | Example for `1.22.3-alpine` |
| --- | --- |
| `{{.Major}}` `{{.Minor}}` `{{.Patch}}` | `1` `22` `3` |
| `{{.Prerelease}}` | `alpine` |
| `{{.Metadata}}` | (build metadata after `+`, if any) |
| `{{.Original}}` | `1.22.3-alpine` |
| `{{.String}}` | `1.22.3-alpine` |

For example `tags: ["{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"]` turns version
`1.22.3-alpine` into target `ghcr.io/me/dev-golang:1.22.3-alpine`.

When several templates are given, the built image is tagged with all of them at
once — the common pattern for floating tags:

```yaml
build:
  repo: ghcr.io/me/dev-golang
  tags:
    - "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
    - "{{.Major}}.{{.Minor}}-alpine"
    - "{{.Major}}-alpine"
```

The first tag is the node's canonical id; its absence in the destination marks
the node outdated. When two selected versions render the same tag (e.g. `1.22`
and `1.23` both render `1-alpine`), the newer version wins it; the older version
simply omits that floating tag.

### Build options

The fields below are optional and shared by both `build` and `bake` kinds (they
map to `docker buildx` options). Paths are relative to the port directory.

| Field | Maps to | Notes |
| --- | --- | --- |
| `dockerfile` | `-f` | Default `Dockerfile`. |
| `context` | build context | Default `.` (the port directory). |
| `target` | `--target` | Dockerfile stage. |
| `platforms` | `--platform` | e.g. `[linux/amd64, linux/arm64]`. |
| `args` | `--build-arg` | `BASE_TAG` (selected tag) is injected for all sources; `BASE` (full reference) for `container` sources. |
| `labels` | `--label` | Base name/digest labels are injected automatically when there is a base. |
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

## The `BASE` and `BASE_TAG` arguments

`clade` injects the **selected upstream tag** as the `BASE_TAG` build argument
for *every* source kind, so the Dockerfile can pin to the exact version:

```dockerfile
ARG BASE_TAG
RUN curl -fsSLo /usr/local/bin/tool \
    "https://example.com/tool@${BASE_TAG}/$(uname -m)"  # BASE_TAG = e.g. 1.2.3
```

For a `container` source, `clade` additionally injects the resolved upstream
*reference* (`repo:tag`) as the `BASE` build argument, so the Dockerfile builds
*on top of the tracked upstream*:

```dockerfile
ARG BASE
FROM ${BASE}
# e.g. BASE = docker.io/library/golang:1.22.3-alpine, BASE_TAG = 1.22.3-alpine
RUN go version
```

Such a build is also labelled automatically:

- `org.opencontainers.image.base.name` — the upstream reference.
- `org.opencontainers.image.base.digest` — the upstream digest (used by the
  `digest` outdated strategy).

> **`http` sources receive `BASE_TAG` but no `BASE`.** They have no upstream
> image, so the Dockerfile declares its own `FROM` and downloads the artifact for
> `${BASE_TAG}`.

## `compare`

How a target is judged outdated when its primary tag already exists. It is an
**ordered list** of strategies tried with fallback: the first that can render a
verdict wins; if a strategy cannot judge the operands (a missing capability) the
next is tried. When omitted, a default is chosen from `source.kind`.

```yaml
compare:
  - kind: digest   # precise: compare the recorded base digest with the base's current digest
  - kind: created  # fallback: compare creation timestamps
```

| `kind` | Outdated when |
| --- | --- |
| `created` | the target was created *before* its base image. |
| `digest` | the base-digest label recorded on the target differs from the base's current digest. `label` (optional) overrides the label key. |

Defaults by `source.kind`:

| Source kind | Default chain |
| --- | --- |
| `container` | `[created, digest]` — timestamp comparison, the same behavior as before per-port config. |
| `http` | *(empty)* — existence only: an existing primary tag is up to date; a new version is detected as a missing primary tag. |

A missing primary tag always marks a node outdated, before any comparator runs.
If every strategy in a non-empty chain is inapplicable, the build aborts (a
configuration error) rather than silently never rebuilding.

## Chaining ports

When a `container` port's `source.repo` equals the `build.repo` of another port,
an **internal edge** is created: the downstream port tracks the tags produced by
the upstream port instead of a registry, and `clade` builds them in order.

```yaml
# ports/base/port.yaml
source: { kind: container, repo: docker.io/library/debian }
select: { kind: semver, last-major: 1 }
build:  { kind: build, repo: ghcr.io/me/base, tags: ["{{.Major}}.{{.Minor}}"] }
```

```yaml
# ports/app/port.yaml  — built on top of ghcr.io/me/base
source: { kind: container, repo: ghcr.io/me/base }
select: { kind: semver }
build:  { kind: build, repo: ghcr.io/me/app, tags: ["{{.Major}}.{{.Minor}}"] }
```

If `ghcr.io/me/base` is rebuilt, every descendant (`ghcr.io/me/app`, ...) is
considered outdated and rebuilt on top of the fresh base. Only `container`
sources chain; an `http` source never forms an internal edge.
