# syntax=docker/dockerfile:1
# Code generated from Dockerfile.tmpl by scripts/gen-dockerfiles; DO NOT EDIT.

ARG BASE=docker.io/library/golang
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
		eza \
		fd-find \
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
		mtr-tiny \
		ncdu \
		net-tools \
		netcat-openbsd \
		openssh-client \
		procps \
		psmisc \
		python3 \
		python3-pip \
		ripgrep \
		rsync \
		socat \
		strace \
		sudo \
		tcpdump \
		telnet \
		tmux \
		traceroute \
		tzdata \
		unzip \
		vim \
		wget \
		whois \
		zip \
		zlib1g \
		zsh \
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
	&& apt-get install -y --no-install-recommends \
		nodejs

ARG USERNAME=hypnos
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd "${USERNAME}" --gid ${USER_GID} \
	&& useradd --create-home "${USERNAME}" --shell /usr/bin/zsh --uid "${USER_UID}" --gid "${USER_GID}" \
	&& echo "${USERNAME}:${USERNAME}" | chpasswd \
	&& echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"

USER ${USERNAME}
WORKDIR /home/${USERNAME}

RUN mkdir -p ~/.cache/go-build

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

USER root
WORKDIR /

# Mount point for a persisted command-history volume. Created with USERNAME
# ownership so a fresh named volume inherits it on first mount.
RUN install -d -o "${USERNAME}" -g "${USERNAME}" /commandhistory

RUN mkdir -p /go/pkg/mod /go/bin \
	&& chown -R "${USERNAME}:${USERNAME}" /go

RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/go/pkg/sumdb \
	for pkg in \
		golang.org/x/tools/gopls@latest \
		github.com/go-delve/delve/cmd/dlv@latest \
		github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
	; do \
		go install "$pkg"; \
	done \
		&& chown -R "${USERNAME}:${USERNAME}" /go/bin
