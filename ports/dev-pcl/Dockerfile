ARG BASE
FROM ${BASE}

RUN apt-get update \
	&& apt-get install --yes --no-install-recommends \
		libxmu-dev \
		libxi-dev \
		libxt-dev \
		libgl-dev \
		libgl1-mesa-dev \
	&& apt-get clean --yes \
	&& rm -rf /var/lib/apt/lists/*



ARG USERNAME=hypnos
ARG VCPKG_REF
RUN cd "${VCPKG_ROOT}" \
	; if [ -n "${VCPKG_REF}" ]; then \
		git fetch --depth=1 origin "${VCPKG_REF}" \
		&& git checkout FETCH_HEAD \
	; fi \
	; sudo -u "${USERNAME}" VCPKG_FORCE_SYSTEM_BINARIES=1 vcpkg install "pcl[visualization]" \
	&& rm -rf \
		"${VCPKG_ROOT}/buildtrees" \
		"${VCPKG_ROOT}/packages" \
		"${VCPKG_DOWNLOADS}"/*
