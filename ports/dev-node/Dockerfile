# syntax=docker/dockerfile:1
# Code generated from Dockerfile.tmpl by scripts/gen-dockerfiles; DO NOT EDIT.

ARG BASE=docker.io/library/node
FROM ${BASE}

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	rm -f /etc/apt/apt.conf.d/docker-clean \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \
		apt-transport-https \
		bash \
		bash-completion \
		build-essential \
		ca-certificates \
		curl \
		dirmngr \
		dnsutils \
		expect \
		fd-find \
		file \
		gettext-base \
		git \
		git-lfs \
		gnupg2 \
		htop \
		init-system-helpers \
		iproute2 \
		iputils-ping \
		jq \
		less \
		libc6 \
		libgcc1 \
		libgssapi-krb5-2 \
		libicu[0-9][0-9] \
		libkrb5-3 \
		libstdc++6 \
		locales \
		lsb-release \
		lsof \
		make \
		man-db \
		manpages \
		manpages-dev \
		moreutils \
		mtr-tiny \
		ncdu \
		net-tools \
		netcat-openbsd \
		openssh-client \
		patch \
		procps \
		psmisc \
		python3 \
		python3-pip \
		ripgrep \
		rsync \
		socat \
		sshpass \
		strace \
		sudo \
		tcpdump \
		telnet \
		tmux \
		traceroute \
		tree \
		tzdata \
		unzip \
		vim \
		wget \
		whois \
		xz-utils \
		zip \
		zlib1g \
		zsh \
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen

# The official node image ships a `node` user at UID/GID 1000, which collides
# with the dev user the shared "user" partial creates at 1000:1000. Remove it
# (and its home) so that partial can claim those IDs.
RUN userdel --remove node 2>/dev/null || true \
	&& groupdel node 2>/dev/null || true

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

RUN git clone --depth 1 https://github.com/ohmyzsh/ohmyzsh.git ~/.oh-my-zsh \
	&& git clone --depth 1 https://github.com/zsh-users/zsh-autosuggestions     ~/.oh-my-zsh/custom/plugins/zsh-autosuggestions \
	&& git clone --depth 1 https://github.com/zsh-users/zsh-syntax-highlighting ~/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting \
	&& git clone --depth 1 https://github.com/zsh-users/zsh-completions         ~/.oh-my-zsh/custom/plugins/zsh-completions

# .zshrc is inlined here (single source of truth across ports). Persists command
# history across rebuilds via the /commandhistory volume (see docker-compose.yaml).
COPY --chown=${USERNAME}:${USERNAME} <<"EOF" /home/${USERNAME}/.zshrc
export ZSH="$HOME/.oh-my-zsh"

ZSH_THEME="refined"

plugins=(
	git
	zsh-autosuggestions
	zsh-syntax-highlighting
)

fpath+=("${ZSH_CUSTOM:-$ZSH/custom}/plugins/zsh-completions/src")

source "$ZSH/oh-my-zsh.sh"

HISTFILE=/commandhistory/.zsh_history
HISTSIZE=100000
SAVEHIST=100000
setopt INC_APPEND_HISTORY
setopt SHARE_HISTORY
setopt HIST_IGNORE_DUPS
EOF

# Node developer tooling, installed globally into a per-user npm prefix so it
# lands under the dev user's home (no root-owned /usr/local writes) and stays on
# PATH. corepack is installed explicitly (recent node images no longer bundle it)
# and provides the pnpm/yarn shims.
ENV NPM_CONFIG_PREFIX=/home/${USERNAME}/.local
ENV PATH=/home/${USERNAME}/.local/bin:${PATH}
# Corepack fetches a package manager on first use and, by default, prompts for
# confirmation before downloading; disable the prompt so pnpm runs
# non-interactively inside the container.
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN npm install -g \
		corepack \
		typescript \
		ts-node \
		npm-check-updates \
	&& corepack enable --install-directory /home/${USERNAME}/.local/bin \
	&& corepack prepare pnpm@latest --activate

USER root
WORKDIR /

# Mount point for a persisted command-history volume. Created with USERNAME
# ownership so a fresh named volume inherits it on first mount.
RUN install -d -o "${USERNAME}" -g "${USERNAME}" /commandhistory
