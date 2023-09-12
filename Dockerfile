# syntax=docker/dockerfile:1
FROM golang:1.21-bookworm as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o "a" ./cmd/clade



FROM debian:bookworm

COPY --from=docker:24-dind "/usr/local/bin/docker" "/usr/local/bin/docker"
COPY --from=builder "/app/a" "/usr/bin/clade"

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	apt update \
	&& apt-get install --no-install-recommends --yes \
		ca-certificates \
		curl \
		jq \
		git

ENTRYPOINT ["clade"]
