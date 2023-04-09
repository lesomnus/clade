ARG BASE
ARG DEV_GCC
FROM ${BASE} AS emsdk
FROM ${DEV_GCC}

COPY --from=emsdk /emsdk /emsdk
ENV PATH="/emsdk:/emsdk/upstream/emscripten:/emsdk/upstream/bin:${PATH}"
