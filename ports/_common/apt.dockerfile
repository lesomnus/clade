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
{{- if .SystemPython }}
		python3 \
		python3-pip \
{{- end }}
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
