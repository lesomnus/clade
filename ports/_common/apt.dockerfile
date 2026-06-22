{{- /*
The base set of apt packages installed into every port. A port may tailor it
when importing this partial:

  {{ template "apt.dockerfile" dict "Exclude" (list "eza") }}
    drop packages the upstream image's apt does not provide (e.g. the node
    image has no eza), or that the base image already ships (the python image
    already has python3/python3-pip).

  {{ template "apt.dockerfile" dict "Include" (list "neovim") }}
    add extra packages on top of the base set.

The final list is sorted so the rendered Dockerfile is deterministic.
*/ -}}
{{- $packages := list
	"apt-transport-https"
	"bash"
	"bash-completion"
	"build-essential"
	"ca-certificates"
	"curl"
	"dirmngr"
	"eza"
	"fd-find"
	"git"
	"git-lfs"
	"gnupg2"
	"htop"
	"init-system-helpers"
	"iproute2"
	"jq"
	"less"
	"libc6"
	"libgcc1"
	"libgssapi-krb5-2"
	"libicu[0-9][0-9]"
	"libkrb5-3"
	"libstdc++6"
	"locales"
	"lsb-release"
	"lsof"
	"make"
	"man-db"
	"manpages"
	"manpages-dev"
	"ncdu"
	"openssh-client"
	"procps"
	"psmisc"
	"python3"
	"python3-pip"
	"ripgrep"
	"rsync"
	"strace"
	"sudo"
	"tmux"
	"tzdata"
	"unzip"
	"vim"
	"wget"
	"zip"
	"zlib1g"
	"zsh"
-}}
{{- $packages = sortStrings (concat (without $packages (optList . "Exclude")) (optList . "Include")) -}}
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	rm -f /etc/apt/apt.conf.d/docker-clean \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends \{{ range $packages }}
		{{ . }} \{{ end }}
	&& echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
	&& locale-gen
