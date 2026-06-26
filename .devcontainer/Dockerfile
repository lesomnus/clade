FROM ghcr.io/lesomnus/dev-golang:1.26

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	install -m 0755 -d /etc/apt/keyrings \
	&& curl -fsSL https://download.docker.com/linux/debian/gpg \
		| gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
	&& chmod a+r /etc/apt/keyrings/docker.gpg \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian bookworm stable" \
		> /etc/apt/sources.list.d/docker.list \
	&& apt update \
	&& apt-get install --yes --no-install-recommends \
		docker-ce-cli

RUN VERSION="1.66" OUT=/usr/local/bin/buf \
	&& curl -fsSL -o "${OUT}" \
		"https://pkg.opt.td/bufbuild/buf/buf@${VERSION}/linux/$(uname -m)" \
	&& chmod +x "${OUT}"


USER hypnos
ENV SHELL=/bin/zsh
