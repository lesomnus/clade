# BASE is injected by `clade build` with the resolved upstream reference,
# e.g. docker.io/library/golang:1.22.3-alpine.
ARG BASE
FROM ${BASE}

ARG EXTRA_TOOLS=""
RUN if [ -n "${EXTRA_TOOLS}" ]; then apk add --no-cache ${EXTRA_TOOLS}; fi

RUN go version
