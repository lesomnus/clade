# syntax=docker/dockerfile:1
# Code generated from Dockerfile.tmpl by scripts/gen-dockerfiles; DO NOT EDIT.

# syntax=docker/dockerfile:1

# This port tracks the official Ubuntu image: it has a `container` source, so
# clade resolves the selected version and injects the full image reference as the
# BASE build-arg, and a rebuild is triggered whenever a newer Ubuntu release is
# published upstream.
ARG BASE=docker.io/library/ubuntu
FROM ${BASE}

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	rm -f /etc/apt/apt.conf.d/docker-clean \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \
		build-essential \
		ca-certificates \
		curl \
		expect \
		fd-find \
		file \
		gettext-base \
		git \
		gnupg \
		jq \
		less \
		locales \
		make \
		moreutils \
		nano \
		openssh-client \
		patch \
		procps \
		python3 \
		python3-pip \
		ripgrep \
		rsync \
		sshpass \
		sudo \
		tree \
		unzip \
		vim \
		wget \
		xz-utils \
		zip \
		zsh \
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen

# Recent Ubuntu images ship an `ubuntu` user at UID/GID 1000, which collides
# with the dev user the shared "user" partial creates at 1000:1000. Remove it
# (and its home) so that partial can claim those IDs.
RUN userdel --remove ubuntu 2>/dev/null || true \
	&& groupdel ubuntu 2>/dev/null || true

# Docker client (the CLI only — no daemon) from Docker's official APT
# repository, which carries newer releases than the distro packages and is the
# only source that ships the buildx/compose CLI plugins.
#
# The repository path is per-distribution and per-release, and the ports do not
# share one: the dev ports sit on Debian base images while the ubuntu port sits
# on Ubuntu. Both the distribution ID and the codename are therefore read from
# /etc/os-release instead of being hardcoded, the same way the LLVM repo is
# resolved in dev-cpp.
#
# Nothing here runs a daemon; a container is expected to reach one on the host
# through a mounted /var/run/docker.sock.
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	. /etc/os-release \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& curl -fsSL "https://download.docker.com/linux/${ID}/gpg" -o /usr/share/keyrings/download.docker.com.asc \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/download.docker.com.asc] https://download.docker.com/linux/${ID} ${VERSION_CODENAME} stable" \
		> /etc/apt/sources.list.d/docker.list \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \
		docker-buildx-plugin \
		docker-ce-cli \
		docker-compose-plugin

ARG USERNAME=hypnos
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd "${USERNAME}" --gid ${USER_GID} \
	&& useradd --create-home "${USERNAME}" --shell /usr/bin/zsh --uid "${USER_UID}" --gid "${USER_GID}" \
	&& echo "${USERNAME}:${USERNAME}" | chpasswd \
	&& echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"

USER ${USERNAME}
WORKDIR /home/${USERNAME}
