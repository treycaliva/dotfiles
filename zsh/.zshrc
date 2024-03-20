#
# Executes commands at the start of an interactive session.
#
# Authors:
#   Sorin Ionescu <sorin.ionescu@gmail.com>
#

# Source Prezto.
if [[ -s "${ZDOTDIR:-$HOME}/.zprezto/init.zsh" ]]; then
  source "${ZDOTDIR:-$HOME}/.zprezto/init.zsh"
fi

# Customize to your needs...
export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES

# export FZF_DEFAULT_COMMAND='rg --files --no-ignore --hidden --follow --glob "!.git/*"'
export FZF_ALT_C_COMMAND='fd -t d . $HOME'
export FZF_DEFAULT_COMMAND='rg --files'
export FZF_DEFAULT_OPTS='-m --height 50% --border'

export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh" # This loads nvm

# source ~/.iterm2_shell_integration.zsh
export PATH="/usr/local/anaconda3/bin:$PATH"
# chruby ruby-2.7.1
# source /usr/local/share/chruby/chruby.sh
# source /usr/local/share/chruby/auto.sh

[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin
# function _update_ps1() {
#   PS1="$($GOPATH/bin/powerline-go -error $?)"
# }
# if [ "$TERM" != "linux" ] && [ -f "$GOPATH/bin/powerline-go" ]; then
#   PROMPT_COMMAND="_update_ps1; $PROMPT_COMMAND"
# fi

autoload -U +X bashcompinit && bashcompinit
complete -o nospace -C /usr/local/bin/terraform terraform

alias fd=fdfind
export PATH="$HOME/.tfenv/bin:$PATH"

# The next line updates PATH for the Google Cloud SDK.
if [ -f '/Users/francis.caliva/google-cloud-sdk/path.zsh.inc' ]; then . '/Users/francis.caliva/google-cloud-sdk/path.zsh.inc'; fi

# The next line enables shell command completion for gcloud.
if [ -f '/Users/francis.caliva/google-cloud-sdk/completion.zsh.inc' ]; then . '/Users/francis.caliva/google-cloud-sdk/completion.zsh.inc'; fi
export PATH="$HOME/.tfenv/bin:$PATH"
