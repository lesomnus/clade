# CLI

```
clade [--config <path>] <command> [flags]
```

`--config` points at a config file; otherwise `clade.yaml` / `clade.yml` in the
working directory is used if present (see [Configuration](#configuration)).

## `clade outdated`

Compute the build graph and print the targets that are out of date with their
upstream.

| Flag | Description |
| --- | --- |
| `--ports <dir>` | Ports directory (default from config, `ports`). |
| `--format <fmt>` | `text` (default), `json`, or `binary`. |
| `--all` | Include up-to-date targets, not just stale ones. |

Output:

- **text** — one line per target: `<status>  <id>  (base: <base>)`.
- **json** — the graph as protojson.
- **binary** — the graph as protobuf wire bytes (pipe or cache it, then feed it
  to `clade build --graph`).

```sh
clade outdated
# outdated  ghcr.io/me/dev-golang:1.24.0-alpine  (base: docker.io/library/golang:1.24-alpine)

clade outdated --format json > graph.json
clade outdated --format binary > graph.pb
```

## `clade build`

Build (and by default push) targets, walking the graph in topological order so a
base is built before anything that depends on it.

```
clade build [node...] [flags]
```

Positional `node` arguments are target references (`repo:tag`) to build. With no
arguments, all **outdated** nodes are built.

| Flag | Description |
| --- | --- |
| `--ports <dir>` | Ports directory (when recomputing the graph). |
| `--graph <file>` | Read a serialized graph (`.json` or binary) instead of recomputing. |
| `--all` | Build every node in the graph, not only outdated ones. |
| `--no-push` | Do not push the built images. |
| `--load` | Load the result into the local image store (implies no push). |
| `--dry-run` | Print the build commands instead of running them. |
| `--docker <bin>` | Binary to invoke (default `docker`). |

A `container`-source build receives the resolved upstream as the `BASE` build
argument and is labelled with `org.opencontainers.image.base.name` and
`org.opencontainers.image.base.digest` (used by the `digest` comparator). An
`http`-source build has no base, so it receives neither.

```sh
clade build                                   # build & push all stale targets
clade build --dry-run                         # preview the buildx commands
clade build ghcr.io/me/dev-golang:1.24.0-alpine   # build one target
clade build --graph graph.pb                  # build from a saved graph
```

## `clade cache`

Inspect and manage the on-disk registry metadata cache (see
[Configuration](#configuration) for where it lives and how long entries live).
The cache stores upstream tag listings and image metadata fetched from
registries, so repeated runs do not re-spend registry rate limit.

```
clade cache ls [repo]        # list cached repositories, or one repo's tags
clade cache rm <repo>...     # drop the cached entries of those repositories
clade cache rm --all         # drop the entire cache
```

`clade cache ls` with no argument lists every repository that has a cached tag
listing, with its tag count and time to expiry:

```sh
clade cache ls
# REPOSITORY                 TAGS  EXPIRES
# docker.io/library/golang     42  in 23h58m12s
# docker.io/library/node       18  expired
```

`clade cache ls <repo>` prints that repository's cached tags, one per line
(sorted), which is convenient to pipe. Removing a repository drops both its tag
listing and any per-tag image metadata cached for it.

| Command | Description |
| --- | --- |
| `cache ls` | List cached repositories (tag count + expiry). |
| `cache ls <repo>` | Print the cached tags of `<repo>`, one per line. |
| `cache rm <repo>...` | Remove the cached entries of the given repositories. |
| `cache rm --all` | Remove every cached entry. |

Expired entries are evicted lazily on the next lookup; `cache ls` still shows
them (marked `expired`) until then, and `cache rm` clears them immediately.

## `clade config`

Print the effective configuration as YAML (defaults merged with the loaded file).

## `clade version`

Print version and build information.

## Configuration

`clade.yaml` (searched in the working directory, or set with `--config`):

```yaml
# Directory that holds port definitions.
ports: ports

# Registry metadata cache (a metadata lookup costs registry rate limit).
cache:
  dir: ""    # default: <user cache dir>/clade
  ttl: 24h   # how long tag listings and image metadata are reused

# Build settings. The build strategy itself is per port (build.kind in port.yaml).
build:
  docker: docker   # docker binary to invoke
```

Outdated comparison is configured **per port** by the `compare` list in
`port.yaml` (an ordered fallback chain), with a default chosen from `source.kind`
when omitted — see [`port.yaml` › `compare`](port.md#compare). A missing primary
tag is always outdated, and a target whose internal base is outdated is rebuilt
as well.

OpenTelemetry can be configured under an `otel:` key; logs go to stderr so
`stdout` stays clean for `--format json|binary`.
