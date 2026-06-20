export ZSH="$HOME/.oh-my-zsh"

ZSH_THEME="refined"

plugins=(
	git
	zsh-autosuggestions
	zsh-syntax-highlighting
)

fpath+=("${ZSH_CUSTOM:-$ZSH/custom}/plugins/zsh-completions/src")

source "$ZSH/oh-my-zsh.sh"

# Persist command history across devcontainer rebuilds via a named volume
# mounted at /commandhistory (see docker-compose.yaml).
HISTFILE=/commandhistory/.zsh_history
HISTSIZE=100000
SAVEHIST=100000
setopt INC_APPEND_HISTORY  # write commands as they are entered, not at shell exit
setopt SHARE_HISTORY       # share history across concurrent sessions
setopt HIST_IGNORE_DUPS
