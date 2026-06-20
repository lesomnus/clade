# Architecture

`clade` turns a set of *ports* into a dependency graph of build targets, marks
which targets are out of date with their upstream, and builds the stale ones in
order.

## Pipeline

```
ports/*/port.yaml ──load──▶ []port.Port
        │
        ├─ for each port: list source versions ──select──▶ render build tag
        │     (container: registry / internal parent │ http: a version endpoint)
        ▼
   graph.Builder ──▶ pb.Graph        (topologically ordered nodes)
        │   └─ fetch target/base metadata, mark outdated (+ propagate to children)
        │
   ┌────┴─────────────────────────┐
   ▼                              ▼
clade outdated                 clade build
(filter + serialize)      (per node: port.yaml → builder.New → Build)
```

## Packages

| Package | Responsibility |
| --- | --- |
| `port` | Parse `port.yaml` (`source`, `select`, `compare`, `build`). Strategy-specific fields are kept as raw `Params` so this package stays free of any source/selector/comparator/builder. |
| `registry` | `Registry` interface (`Tags`, `Stat`) + `Remote` (go-containerregistry), a TTL cache decorator (`WithCache`, mem/file), and an in-memory `Fake`. |
| `source` | `Source` interface (`Versions`) to discover upstream versions, with a kind registry. `container` (lists registry tags via an injected lister) and `http` (fetches a version string) are built in. |
| `tag` | `Selector` interface to select among versions, with a kind registry. `semver` is the built-in strategy (and the parser feeding the build-tag templates). |
| `compare` | `Comparator` over a sealed, opaque `Comparable` inspected through capability interfaces (`Created`, `Digested`, `Labeled`); `created` and `digest` built in and composed into a fallback `Chain`. Configured per port. |
| `graph` | `Builder` expands ports into concrete nodes, topologically sorts them, fetches metadata, and marks outdated nodes (propagating to descendants). |
| `builder` | `Builder` interface (`Build(ctx)`) with a kind registry. `build` (`docker buildx build`) and `bake` (`docker buildx bake`) are built in. |
| `pb/clade/v1` | Generated graph types (`Image`, `Node`, `Graph`). Source: `proto/clade/v1/graph.proto`. |
| `cmd`, `cmd/config`, `cmd/version` | CLI wiring (built on `xli`) and configuration. |

## Pluggable abstractions

Four concerns are factored behind interfaces with a `kind → factory` registry,
so new strategies are added with a `Register` call and a small implementation:

- **Version discovery** (`source.Source`) — where upstream versions come from.
  Selected by `source.kind` in `port.yaml` (`container`, `http`).
- **Version selection** (`tag.Selector`) — which versions to track. Selected by
  `select.kind` in `port.yaml`.
- **Outdated check** (`compare.Comparator`) — how to decide a target is stale.
  Configured per port by the `compare` list (an ordered fallback chain), or the
  default for the port's `source.kind` when omitted.
- **Build backend** (`builder.Builder`) — how to build an image. Selected by
  `build.kind` in `port.yaml`.

A `builder.Builder` is constructed from two inputs and then just runs:

- `params` — the raw `build` YAML of the port (strategy-specific options).
- `builder.Spec` — the universal runtime description of one build (port dir,
  tags, base reference, injected labels, push/load, dry-run, output). This keeps
  the per-strategy options out of any shared struct.

## The graph

`graph.Builder.Build` produces a `pb.Graph`:

- **Nodes** are concrete target images (`repo:tag`). Each carries its `base`
  reference, the producing `port` directory, internal `parents`, and an
  `outdated` flag. *How* to build (Dockerfile, context, buildx options) is **not**
  in the node — it is read back from the port's `port.yaml` at build time.
- **Edges** connect an internal parent (one of your ports) to its dependents.
  External upstreams (e.g. `docker.io/library/golang`) have no node.
- Nodes are ordered topologically, so parents are always built before children.

A node is outdated when its primary tag is missing, when its comparator chain
reports it stale relative to its base, or when any internal ancestor is
outdated. A target with an empty chain (e.g. an `http` source, which has no base
image) is judged by existence only: an existing primary tag is up to date.

## Caching

`registry.WithCache` wraps a `Registry` so tag listings and image metadata are
reused for a TTL (`cache.ttl`, default 24h) from a memory or filesystem store
(`cache.dir`, default `<user cache dir>/clade`). The build step resolves a base
image's digest with a *fresh* (uncached) registry, so a base that was just
rebuilt and pushed is reflected immediately.

Entries are keyed `tags:<repo>` and `stat:<ref>` (the `registry.KeyTags` /
`registry.KeyStat` prefixes). The `FileCache` stores each entry under its key's
hash but records the key inside the file, so its key is recoverable for
inspection; that backs the `clade cache` command (`Entries`/`Remove`/`Clear`),
which lists cached repositories, prints a repository's cached tags, and evicts
entries on demand.

## Build automation

`.github/workflows/refresh.yaml` (cron) builds `clade`, runs `clade outdated`,
caches the graph, and dispatches `.github/workflows/build.yaml` once per stale
node — sequentially, so an internal base is pushed before its dependents build
on it. Each build job runs `clade build <node>`.
