ARG BASE
FROM ${BASE}

RUN apt-get update \
	&& apt-get install --yes --no-install-recommends \
		apt-utils \
		openssh-client \
		gnupg2 \
		dirmngr \
		iproute2 \
		procps \
		lsof \
		htop \
		net-tools \
		psmisc \
		curl \
		wget \
		rsync \
		ca-certificates \
		unzip \
		zip \
		git \
		git-lfs \
		nano \
		vim \
		less \
		jq \
		lsb-release \
		apt-transport-https \
		dialog \
		libc6 \
		libgcc1 \
		libkrb5-3 \
		libgssapi-krb5-2 \
		libicu[0-9][0-9] \
		liblttng-ust0 \
		libstdc++6 \
		zlib1g \
		locales \
		sudo \
		ncdu \
		man-db \
		strace \
		manpages \
		manpages-dev \
		init-system-helpers \
		bash-completion \
		zsh \
	&& apt-get clean --yes \
	&& rm -rf /var/lib/apt/lists/* \
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen

RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
	&& apt-get install --yes --no-install-recommends \
		nodejs \
	&& apt-get clean --yes \
	&& rm -rf /var/lib/apt/lists/* 



ARG USERNAME=hypnos
ARG USER_UID=${UID:-1000}
ARG USER_GID=${GID:-${USER_UID}}
RUN groupadd "${USERNAME}" --gid ${USER_GID} \
	&& useradd --create-home "${USERNAME}" --shell /bin/bash --uid "${USER_UID}" --gid "${USER_GID}" \
	&& echo "${USERNAME}:${USERNAME}" | chpasswd \
	&& echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"

WORKDIR /home/${USERNAME}
