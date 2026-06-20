# syntax=docker/dockerfile:1
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
		eza \
		fd-find \
		git \
		git-lfs \
		gnupg2 \
		htop \
		init-system-helpers \
		iproute2 \
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
		ncdu \
		openssh-client \
		procps \
		psmisc \
		python3 \
		python3-pip \
		ripgrep \
		rsync \
		strace \
		sudo \
		tmux \
		tzdata \
		unzip \
		vim \
		wget \
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
COPY --chown=${USERNAME}:${USERNAME} .zshrc /home/${USERNAME}/.zshrc


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
