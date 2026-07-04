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
		fd-find \
		file \
		git \
		gnupg \
		jq \
		less \
		locales \
		make \
		nano \
		openssh-client \
		procps \
		python3 \
		python3-pip \
		ripgrep \
		rsync \
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

ARG USERNAME=hypnos
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd "${USERNAME}" --gid ${USER_GID} \
	&& useradd --create-home "${USERNAME}" --shell /usr/bin/zsh --uid "${USER_UID}" --gid "${USER_GID}" \
	&& echo "${USERNAME}:${USERNAME}" | chpasswd \
	&& echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"

USER ${USERNAME}
WORKDIR /home/${USERNAME}
