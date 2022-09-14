ARG TAG
ARG BASE
FROM ${BASE}:${TAG}

RUN apt-get update \
	&& apt-get install --yes --no-install-recommends \
		bison \
		nasm \
	&& apt-get clean --yes \
	&& rm -rf /var/lib/apt/lists/*



ARG USERNAME=hypnos
ARG VCPKG_REF
RUN cd "${VCPKG_ROOT}" \
	; if [ -n "$VCPKG_REF" ]; then \
		git fetch --depth=1 origin "${VCPKG_REF}" \
		&& git checkout FETCH_HEAD \
	; fi \
	; sudo -u ${USERNAME} vcpkg install "ffmpeg[ \
		aom, \
		ilbc, \
		openh264, \
		openjpeg, \
		webp, \
		x264, \
		x265, \
		zlib, \
		nvcodec, \
		opencl, \
		ffmpeg, \
		ffplay, \
		ffprobe \
	]" \
	&& rm -rf \
		"${VCPKG_ROOT}/buildtrees" \
		"${VCPKG_ROOT}/packages" \
		"${VCPKG_DOWNLOADS}"/*
