FROM ghcr.io/lesomnus/dev-docker:24

COPY --from=registry.hub.docker.com/library/golang:1.21 \
	/usr/local/go \
	/usr/local/go

ENV GOPATH=/home/hypnos/go
ENV PATH $GOPATH/bin:$PATH



RUN apt-get update \
	&& apt-get install --yes --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
	&& apt-get clean --yes \
	&& rm -rf /var/lib/apt/lists/*



USER hypnos

# Comment out the following lines if you don't want to use Zsh.
RUN wget https://github.com/robbyrussell/oh-my-zsh/raw/master/tools/install.sh -O - | zsh || true \
	&& git clone https://github.com/zsh-users/zsh-autosuggestions         ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions     \
	&& git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting \
	&& git clone https://github.com/zsh-users/zsh-completions             ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-completions         \
	&& git clone https://github.com/supercrabtree/k                       ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/k
ADD .zshrc /home/hypnos
ENV SHELL /bin/zsh
