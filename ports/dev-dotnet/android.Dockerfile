ARG TAG
ARG BASE
FROM ${BASE}:${TAG}

RUN dotnet workload install android
# TODO
