# syntax=docker/dockerfile:1

# Claude Code ships as a single binary, so this port has an `http` source rather
# than a parent image: it has no base image (no FROM ${BASE}), so the base is
# chosen here. clade injects the selected version as the BASE_TAG build-arg, and
# a rebuild is triggered whenever the stable release advances (its primary
# {{.Major}}.{{.Minor}}.{{.Patch}} tag goes missing in the destination repo).
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

ARG BASE_TAG
RUN OUT=/usr/local/bin/claude \
	&& curl -fsSL -o "${OUT}" \
		"https://pkg.opt.td/claude.ai/code@${BASE_TAG}/linux/$(uname -m)" \
	&& chmod +x "${OUT}" \
	&& claude --version

ENTRYPOINT ["claude"]
