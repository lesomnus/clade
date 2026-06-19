FROM ghcr.io/lesomnus/dev-golang:1.26.4

RUN VERSION="2.1.181" OUT=/usr/local/bin/claude \
	&& curl -fsSL -o "${OUT}" \
		"https://pkg.opt.td/claude.ai/code@${VERSION}/linux/$(uname -m)" \
	&& chmod +x "${OUT}"

RUN VERSION="1.66" OUT=/usr/local/bin/buf \
	&& curl -fsSL -o "${OUT}" \
		"https://pkg.opt.td/bufbuild/buf/buf@${VERSION}/linux/$(uname -m)" \
	&& chmod +x "${OUT}"

USER hypnos
ENV SHELL=/bin/zsh
