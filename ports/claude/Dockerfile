# syntax=docker/dockerfile:1

# Claude Code ships as a single binary, so this port has an `http` source rather
# than a parent image. clade injects no build-arg for http sources, so the base
# image is chosen here and the version is resolved at build time from the same
# endpoint clade tracks. A rebuild is triggered whenever the stable release
# advances (its primary {{.Major}}.{{.Minor}}.{{.Patch}} tag goes missing in the
# destination repository).
FROM docker.io/library/debian:bookworm-slim

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	rm -f /etc/apt/apt.conf.d/docker-clean \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		curl \
		git \
	&& rm -rf /var/lib/apt/lists/*

RUN VERSION="$(curl -fsSL https://downloads.claude.ai/claude-code-releases/stable)" \
	&& OUT=/usr/local/bin/claude \
	&& curl -fsSL -o "${OUT}" \
		"https://pkg.opt.td/claude.ai/code@${VERSION}/linux/$(uname -m)" \
	&& chmod +x "${OUT}" \
	&& claude --version

ENTRYPOINT ["claude"]
