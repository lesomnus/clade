FROM ghcr.io/lesomnus/dev-golang:1.26


USER hypnos

RUN mkdir -p \
		~/go/pkg/mod \
		~/.cache/go-build

# Comment out the following lines if you don't want to use Zsh.
RUN curl -fsSL https://github.com/robbyrussell/oh-my-zsh/raw/master/tools/install.sh | zsh || true \
	&& git clone https://github.com/zsh-users/zsh-autosuggestions         ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions     \
	&& git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting \
	&& git clone https://github.com/zsh-users/zsh-completions             ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-completions
COPY .zshrc /home/hypnos
ENV SHELL=/bin/zsh
