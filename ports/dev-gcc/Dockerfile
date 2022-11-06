ARG TAG
ARG BASE
FROM ${BASE}:${TAG}

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
		python3 \
		pip \
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
RUN groupadd "${USERNAME}" --gid "${USER_GID}" \
	&& useradd --create-home "${USERNAME}" --shell /bin/bash --uid "${USER_UID}" --gid "${USER_GID}" \
	&& echo "${USERNAME}:${USERNAME}" | chpasswd \
	&& echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"

WORKDIR /home/${USERNAME}



RUN apt-get update \
	&& apt-get install --no-install-recommends --yes \
		gdb \
		build-essential \
		pkg-config \
		ninja-build \
		doxygen \
		graphviz \
		gcovr \
		clang-tidy \
	&& apt-get clean --yes \
	&& rm -rf /var/lib/apt/lists/*



# Install CMake
ARG CMAKE_VERSION
RUN ARCH= && dpkgArch="$(dpkg --print-architecture)" \
	&& case "${dpkgArch##*-}" in \
		amd64) ARCH='x86_64';; \
		arm64) ARCH='aarch64';; \
		*) echo "unsupported architecture"; exit 1 ;; \
	esac \
	&& TMP_DIR=$(mktemp -d -t cmake-${CMAKE_VERSION}-XXXX) \
	&& echo "${TMP_DIR}" \
	&& cd "${TMP_DIR}" \
	&& curl -sSL "https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}-linux-${ARCH}.tar.gz" -O \
	&& curl -sSL "https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}-SHA-256.txt" -o checksum \
	&& sha256sum --check --ignore-missing "checksum" \
	&& tar -xzf "cmake-${CMAKE_VERSION}-linux-${ARCH}.tar.gz" \
		--exclude bin/cmake-gui \
		--exclude doc/cmake \
		--exclude "share/cmake-*/Help" \
	&& cd */ \
	&& mkdir -p /opt/cmake \
	&& cp -rf * /opt/cmake/. \
	&& ln -s /opt/cmake/bin/cmake /usr/local/bin/cmake \
	&& ln -s /opt/cmake/bin/ctest /usr/local/bin/ctest \
	&& cmake --version \
	&& rm -rf "${TMP_DIR}"



# Install vcpkg
ENV VCPKG_ROOT=/opt/vcpkg \
	VCPKG_DOWNLOADS=/opt/vcpkg-downloads \
	VCPKG_FORCE_SYSTEM_BINARIES=1
RUN mkdir -p "${VCPKG_ROOT}" "${VCPKG_DOWNLOADS}" \
	&& chown -R "${USER_UID}:${USER_GID}" "${VCPKG_ROOT}" "${VCPKG_DOWNLOADS}" \
	&& sudo -u "${USERNAME}" git clone --single-branch https://github.com/microsoft/vcpkg "${VCPKG_ROOT}" \
	&& cd "${VCPKG_ROOT}" \
	&& ./bootstrap-vcpkg.sh -disableMetrics \
	&& ln -s /opt/vcpkg/vcpkg /usr/local/bin/vcpkg \
	&& sudo -u "${USERNAME}" VCPKG_FORCE_SYSTEM_BINARIES=1 vcpkg integrate bash \
	&& vcpkg --version \
	&& rm -rf \
		"${VCPKG_ROOT}/buildtrees" \
		"${VCPKG_ROOT}/packages" \
		"${VCPKG_DOWNLOADS}"/*
