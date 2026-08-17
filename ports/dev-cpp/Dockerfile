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
		expect \
		fd-find \
		file \
		gdb \
		gettext-base \
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
		moreutils \
		mtr-tiny \
		ncdu \
		net-tools \
		netcat-openbsd \
		ninja-build \
		openssh-client \
		patch \
		pkg-config \
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
		valgrind \
		vim \
		wget \
		whois \
		xz-utils \
		zip \
		zlib1g \
		zsh \
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen

# Clang/LLVM toolchain from apt.llvm.org — the LLVM project's official APT
# repository, which ships a newer toolchain than the base image's bundled clang.
# We add the distribution's unversioned suite (it tracks the newest LLVM line for
# that release) and detect whichever clang version it offers, rather than pinning
# one: an older gcc base image runs on an older Debian whose apt.llvm.org repo may
# not carry the latest line, so a pinned version would 404 there. The detected
# version then selects the matching packages. update-alternatives exposes the
# tools under their unversioned names (clang, clang++, clangd, ...).
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	. /etc/os-release \
	&& curl -fsSL https://apt.llvm.org/llvm-snapshot.gpg.key -o /usr/share/keyrings/apt.llvm.org.asc \
	&& echo "deb [signed-by=/usr/share/keyrings/apt.llvm.org.asc] http://apt.llvm.org/${VERSION_CODENAME}/ llvm-toolchain-${VERSION_CODENAME} main" \
		> /etc/apt/sources.list.d/llvm.list \
	&& apt-get update \
	&& llvm_version="$(apt-cache search --names-only '^clang-[0-9]+$' | awk '{print $1}' | grep -oE '[0-9]+$' | sort -un | tail -1)" \
	&& if [ -z "${llvm_version}" ]; then \
		echo "no clang package found in apt.llvm.org repo for ${VERSION_CODENAME}" >&2; exit 1; \
	fi \
	&& echo "installing LLVM/clang ${llvm_version}" \
	&& apt-get install -y --no-install-recommends \
		clang-${llvm_version} \
		clang-format-${llvm_version} \
		clang-tidy-${llvm_version} \
		clang-tools-${llvm_version} \
		clangd-${llvm_version} \
		lld-${llvm_version} \
		lldb-${llvm_version} \
	&& for tool in \
		clang clang++ clangd clang-format clang-tidy \
		clang-apply-replacements run-clang-tidy lld ld.lld lldb \
	; do \
		update-alternatives --install "/usr/bin/${tool}" "${tool}" "/usr/bin/${tool}-${llvm_version}" 100 || true; \
	done

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
	&& apt-get install -y --no-install-recommends \
		nodejs

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

USER root
WORKDIR /

# Mount point for a persisted command-history volume. Created with USERNAME
# ownership so a fresh named volume inherits it on first mount.
RUN install -d -o "${USERNAME}" -g "${USERNAME}" /commandhistory
