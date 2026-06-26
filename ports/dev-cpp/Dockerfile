# syntax=docker/dockerfile:1
# Code generated from Dockerfile.tmpl by scripts/gen-dockerfiles; DO NOT EDIT.

ARG BASE=docker.io/library/gcc
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
		ccache \
		cmake \
		cppcheck \
		curl \
		dirmngr \
		dnsutils \
		fd-find \
		gdb \
		git \
		git-lfs \
		gnupg2 \
		htop \
		init-system-helpers \
		iproute2 \
		iputils-ping \
		jq \
		lcov \
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
		ninja-build \
		openssh-client \
		pkg-config \
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
		valgrind \
		vim \
		wget \
		whois \
		zip \
		zlib1g \
		zsh \
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen

# Clang/LLVM toolchain from apt.llvm.org — the LLVM project's official APT
# repository. It ships a newer toolchain than Debian trixie's bundled clang-19;
# LLVM_VERSION pins the release line (bump it to upgrade). update-alternatives
# exposes the tools under their unversioned names (clang, clang++, clangd, ...).
ARG LLVM_VERSION=22
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	. /etc/os-release \
	&& curl -fsSL https://apt.llvm.org/llvm-snapshot.gpg.key -o /usr/share/keyrings/apt.llvm.org.asc \
	&& echo "deb [signed-by=/usr/share/keyrings/apt.llvm.org.asc] http://apt.llvm.org/${VERSION_CODENAME}/ llvm-toolchain-${VERSION_CODENAME}-${LLVM_VERSION} main" \
		> /etc/apt/sources.list.d/llvm.list \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \
		clang-${LLVM_VERSION} \
		clang-format-${LLVM_VERSION} \
		clang-tidy-${LLVM_VERSION} \
		clang-tools-${LLVM_VERSION} \
		clangd-${LLVM_VERSION} \
		lld-${LLVM_VERSION} \
		lldb-${LLVM_VERSION} \
	&& for tool in \
		clang clang++ clangd clang-format clang-tidy \
		clang-apply-replacements run-clang-tidy lld ld.lld lldb \
	; do \
		update-alternatives --install "/usr/bin/${tool}" "${tool}" "/usr/bin/${tool}-${LLVM_VERSION}" 100 || true; \
	done

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
