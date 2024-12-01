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

# Add in Powerlevel10k
zinit ice depth=1; zinit light romkatv/powerlevel10k

<<<<<<< HEAD:zsh/zshrc
# Add in zsh plugins
zinit light zsh-users/zsh-autosuggestions
zinit light zsh-users/zsh-completions
zinit light zsh-users/zsh-syntax-highlighting
=======
# source ~/.iterm2_shell_integration.zsh
export PATH="/usr/local/anaconda3/bin:$PATH"
# chruby ruby-2.7.1
# source /usr/local/share/chruby/chruby.sh
# source /usr/local/share/chruby/auto.sh
>>>>>>> 09ccdbfbff7a929d6546490849f2a155153256fc:zsh/.zshrc

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
GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin

# Terraform stuff
autoload -U +X bashcompinit && bashcompinit
complete -o nospace -C /usr/local/bin/terraform terraform
export PATH="$HOME/.tfenv/bin:$PATH"

# GCloud
# The next line updates PATH for the Google Cloud SDK.
if [ -f '/Users/treycaliva/google-cloud-sdk/path.zsh.inc' ]; then . '/Users/treycaliva/google-cloud-sdk/path.zsh.inc'; fi

# The next line enables shell command completion for gcloud.
<<<<<<< HEAD:zsh/zshrc
if [ -f '/Users/treycaliva/google-cloud-sdk/completion.zsh.inc' ]; then . '/Users/treycaliva/google-cloud-sdk/completion.zsh.inc'; fi
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion

eval "$(zoxide init zsh --cmd cd)"

alias k="kubectl"
alias ls="ls --color=auto"
=======
if [ -f '/Users/francis.caliva/google-cloud-sdk/completion.zsh.inc' ]; then . '/Users/francis.caliva/google-cloud-sdk/completion.zsh.inc'; fi
export PATH="$HOME/.tfenv/bin:$PATH"
>>>>>>> 09ccdbfbff7a929d6546490849f2a155153256fc:zsh/.zshrc
