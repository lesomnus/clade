# Architecture

`clade` turns a set of *ports* into a dependency graph of build targets, marks
which targets are out of date with their upstream, and builds the stale ones in
order.

## Pipeline

```
ports/*/port.yaml ──load──▶ []port.Port
        │
        ├─ for each port: list parent tags ──select──▶ render build tag
        │     (external: registry │ internal: parent port's produced tags)
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
| `port` | Parse `port.yaml`. `Build.Params` keeps the raw build config so this package stays free of any builder. |
| `registry` | `Registry` interface (`Tags`, `Stat`) + `Remote` (go-containerregistry), a TTL cache decorator (`WithCache`, mem/file), and an in-memory `Fake`. |
| `tag` | `Selector` interface to choose upstream tags, with a kind registry. `semver` is the built-in strategy. |
| `compare` | `Comparator` interface deciding if a target is outdated, with a kind registry. `created` and `digest` are built in. |
| `graph` | `Builder` expands ports into concrete nodes, topologically sorts them, fetches metadata, and marks outdated nodes (propagating to descendants). |
| `builder` | `Builder` interface (`Build(ctx)`) with a kind registry. `build` (`docker buildx build`) and `bake` (`docker buildx bake`) are built in. |
| `pb/clade/v1` | Generated graph types (`Image`, `Node`, `Graph`). Source: `proto/clade/v1/graph.proto`. |
| `cmd`, `cmd/config`, `cmd/version` | CLI wiring (built on `xli`) and configuration. |

## Pluggable abstractions

Three concerns are factored behind interfaces with a `kind → factory` registry,
so new strategies are added with a `Register` call and a small implementation:

- **Tag selection** (`tag.Selector`) — which upstream tags to track. Selected by
  `parent.target.kind` in `port.yaml`.
- **Outdated check** (`compare.Comparator`) — how to decide a target is stale.
  Selected by `compare.kind` in `clade.yaml`.
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

A node is outdated when its target image is missing, when the comparator reports
it older than its base, or when any internal ancestor is outdated.

## Caching

`registry.WithCache` wraps a `Registry` so tag listings and image metadata are
reused for a TTL (`cache.ttl`, default 24h) from a memory or filesystem store
(`cache.dir`, default `<user cache dir>/clade`). The build step resolves a base
image's digest with a *fresh* (uncached) registry, so a base that was just
rebuilt and pushed is reflected immediately.

## Build automation

`.github/workflows/refresh.yaml` (cron) builds `clade`, runs `clade outdated`,
caches the graph, and dispatches `.github/workflows/build.yaml` once per stale
node — sequentially, so an internal base is pushed before its dependents build
on it. Each build job runs `clade build <node>`.
