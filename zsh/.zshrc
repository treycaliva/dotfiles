# Enable Powerlevel10k instant prompt. Should stay close to the top of ~/.zshrc.
# Initialization code that may require console input (password prompts, [y/n]
# confirmations, etc.) must go above this block; everything else may go below.
if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

# Set the directory we want to store zinit and plugins
ZINIT_HOME="${XDG_DATA_HOME:-${HOME}/.local/share}/zinit/zinit.git"
[ ! -d $ZINIT_HOME ] && mkdir -p "$(dirname $ZINIT_HOME)"
[ ! -d $ZINIT_HOME/.git ] && git clone https://github.com/zdharma-continuum/zinit.git "$ZINIT_HOME"

# Source/Load zinit
source "${ZINIT_HOME}/zinit.zsh"

if [[ "$TERM_PROGRAM" == "ghostty" ]]; then
  export TERM=xterm-256color
fi

# Add in Powerlevel10k
zinit ice depth=1; zinit light romkatv/powerlevel10k

# Add in zsh plugins
zinit light zsh-users/zsh-autosuggestions
zinit light zsh-users/zsh-completions
zinit light zsh-users/zsh-syntax-highlighting
# source ~/.iterm2_shell_integration.zsh
export PATH="/usr/local/anaconda3/bin:$PATH"
# chruby ruby-2.7.1
# source /usr/local/share/chruby/chruby.sh
# source /usr/local/share/chruby/auto.sh

# history-substring-search
zinit snippet OMZ::plugins/git/git.plugin.zsh
zinit load zsh-users/zsh-history-substring-search
zinit ice wait atload'_history_substring_search_config'

# Load completions
autoload -U compinit && compinit

# Key bindings
bindkey '^I' complete-word
bindkey '^[[Z' autosuggest-accept
bindkey '^[[A' history-substring-search-up
bindkey '^[[B' history-substring-search-down

# To customize prompt, run `p10k configure` or edit ~/.p10k.zsh.
[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh

# History
HISTSIZE=10000
HISTFILE=~/.zsh_history
SAVEHIST=$HISTSIZE
HISTDUP=erase
setopt appendhistory
setopt sharehistory
setopt hist_ignore_space
setopt hist_ignore_all_dups
setopt hist_save_no_dups
setopt hist_ignore_dups
setopt hist_find_no_dups

# Go stuff
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Terraform stuff
autoload -U +X bashcompinit && bashcompinit
complete -o nospace -C /usr/local/bin/terraform terraform
export PATH="$HOME/.tfenv/bin:$PATH"

# GCloud[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion

eval "$(zoxide init zsh --cmd cd)"

[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

autoload -U +X bashcompinit && bashcompinit
complete -o nospace -C /usr/local/bin/terraform terraform

alias fd=fdfind
# Load a few important annexes, without Turbo
# (this is currently required for annexes)
zinit light-mode for \
    zdharma-continuum/zinit-annex-as-monitor \
    zdharma-continuum/zinit-annex-bin-gem-node \
    zdharma-continuum/zinit-annex-patch-dl \
    zdharma-continuum/zinit-annex-rust

### End of Zinit's installer chunk

# The next line updates PATH for the Google Cloud SDK.
if [ -f '/Users/treycaliva/google-cloud-sdk/path.zsh.inc' ]; then . '/Users/treycaliva/google-cloud-sdk/path.zsh.inc'; fi

# The next line enables shell command completion for gcloud.
if [ -f '/Users/treycaliva/google-cloud-sdk/completion.zsh.inc' ]; then . '/Users/treycaliva/google-cloud-sdk/completion.zsh.inc'; fi

EDITOR="nvim"
# Alias
alias k="kubectl"
alias ls="ls --color=auto"
source <(kubectl completion zsh)

# export NVM_DIR="$HOME/.nvm"
# [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
export NVM_LAZY_LOAD=true
source "/Users/treycaliva/.zsh-nvm.zsh"

export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init -)"

. "$HOME/.local/bin/env"

gcup() {
  # 1. Get the name of the branch you're currently on
  local current_branch
  current_branch=$(git rev-parse --abbrev-ref HEAD)

  # 2. Safety check: Don't try to delete main or master
  if [ "$current_branch" = "main" ] || [ "$current_branch" = "master" ]; then
    echo "You are already on '$current_branch'. Running 'git pull'."
    git pull
  else
    # 3. This is your desired command sequence
    echo "Switching to main, pulling, and deleting local branch '$current_branch'..."
    git checkout main && git pull && git branch -d "$current_branch"
  fi
}
export PATH=~/.groundcover/bin:/$PATH

# Added by Antigravity
export PATH="/Users/treycaliva/.antigravity/antigravity/bin:$PATH"
export PATH="$HOME/.local/bin:$PATH"
