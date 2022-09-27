ARG TAG
ARG BASE
FROM ${BASE}:${TAG}

RUN apk update \
	&& apk add --no-cache \
		openssh-client \
		gnupg \
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
		libgcc \
		libstdc++ \
		krb5-libs \
		libintl \
		libssl1.1 \
		lttng-ust \
		tzdata \
		userspace-rcu \
		zlib \
		sudo \
		coreutils \
		sed \
		grep \
		which \
		ncdu \
		shadow \
		strace \
		bash-completion \
		zsh \
		python3 \
		py3-pip \
		nodejs \
		npm



ARG USERNAME=hypnos
ARG USER_UID=${UID:-1000}
ARG USER_GID=${GID:-${USER_UID}}
RUN groupadd ${USERNAME} --gid ${USER_GID} \
	&& useradd --create-home ${USERNAME} --shell /bin/bash --uid ${USER_UID} --gid ${USER_GID} \
	&& echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"

WORKDIR /home/${USERNAME}
