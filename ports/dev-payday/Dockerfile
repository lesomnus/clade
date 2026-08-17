# syntax=docker/dockerfile:1
# Code generated from Dockerfile.tmpl by scripts/gen-dockerfiles; DO NOT EDIT.

ARG BASE=ghcr.io/lesomnus/dev-golang
FROM ${BASE}

# payday builds on dev-golang rather than on a stock golang image, so everything
# the other dev ports share -- the apt set, node, the docker client, the user and
# the shell -- is already here. What is added below is only what payday itself
# needs and no other port does.

# buf, because payday's sources are `.proto`: its schema, its generators and its
# compatibility checks (`buf breaking` against two baselines) all run through it,
# and a checkout without buf cannot regenerate what it ships.
#
# Fetched through pkg.opt.td, which is arrakis (github.com/lesomnus/arrakis): a
# redirector to the release the project itself published. It is what saves this
# from carrying an arch table -- bufbuild names its assets `Linux-x86_64` and
# `Linux-aarch64`, arrakis maps `$(uname -m)` onto that and 301s to the GitHub
# release, so this layer is one curl on either platform.
#
# The pin is the series, so a rebuild picks up the patch releases of 1.72 but
# never a minor. It is manual either way: this port's version follows the Go
# toolchain it sits on (see port.yaml), so nothing moves buf on its own, and
# payday's CI runs whatever `bufbuild/buf-action` ships, which is the newest.
ARG BUF_VERSION=1.72
RUN curl -fsSL -o /usr/local/bin/buf \
		"https://pkg.opt.td/bufbuild/buf/buf@${BUF_VERSION}/linux/$(uname -m)" \
	&& chmod +x /usr/local/bin/buf \
	&& buf --version

# A browser, so a page can be looked at rather than reasoned about.
#
# Two things want it. payday's sandbox test (`PDTEST_SANDBOX=1`) serves the wasm
# build through vite and drives it with `chromium.launch()`, which asserts the
# one thing neither half's unit tests can: that the Go program and the TypeScript
# program are the same program. And an agent working in this container has no
# other way to answer whether a word fits a button, whether a colour reads on a
# dark background, or which element is giving a phone a sideways scrollbar.
#
# `--with-deps` is why this is a root layer: chromium wants a couple of dozen
# shared libraries a Go image has no reason to carry. The browsers land outside
# any home directory and are made world-readable, so the container's user runs
# them without owning them.
#
# The version is pinned to what payday's `package.json` resolves for playwright.
# The library and the browser talk over a protocol that changes, and a browser
# newer than the library driving it fails in ways that read as the page being
# broken -- so this moves when payday's lockfile does, not on its own.
ARG PLAYWRIGHT_VERSION=1.62.1
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/playwright
RUN npx --yes "playwright@${PLAYWRIGHT_VERSION}" install --with-deps chromium \
	&& chmod -R a+rX "${PLAYWRIGHT_BROWSERS_PATH}" \
	&& rm -rf /root/.npm
