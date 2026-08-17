# Docker client (the CLI only — no daemon) from Docker's official APT
# repository, which carries newer releases than the distro packages and is the
# only source that ships the buildx/compose CLI plugins.
#
# The repository path is per-distribution and per-release, and the ports do not
# share one: the dev ports sit on Debian base images while the ubuntu port sits
# on Ubuntu. Both the distribution ID and the codename are therefore read from
# /etc/os-release instead of being hardcoded, the same way the LLVM repo is
# resolved in dev-cpp.
#
# Nothing here runs a daemon; a container is expected to reach one on the host
# through a mounted /var/run/docker.sock.
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	. /etc/os-release \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& curl -fsSL "https://download.docker.com/linux/${ID}/gpg" -o /usr/share/keyrings/download.docker.com.asc \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/download.docker.com.asc] https://download.docker.com/linux/${ID} ${VERSION_CODENAME} stable" \
		> /etc/apt/sources.list.d/docker.list \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \
		docker-buildx-plugin \
		docker-ce-cli \
		docker-compose-plugin
