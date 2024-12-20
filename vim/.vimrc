let data_dir = has('nvim') ? stdpath('data') . '/site' : '~/.vim'
  if empty(glob(data_dir . '/autoload/plug.vim'))
  silent execute '!curl -fLo '.data_dir.'/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim'
  autocmd VimEnter * PlugInstall --sync | source $MYVIMRC
endif

call plug#begin('~/.vim/plugged')
Plug 'airblade/vim-gitgutter'
Plug 'Akulen/vim-dcrpc'
Plug 'burnettk/vim-jenkins'
Plug 'christoomey/vim-tmux-navigator'
Plug 'christoomey/vim-tmux-runner'
Plug 'dense-analysis/ale'
Plug 'dikiaap/minimalist'
Plug 'dracula/vim', {'as':'dracula'}
Plug 'ekalinin/Dockerfile.vim'
Plug 'fatih/vim-go'
Plug 'fatih/vim-hclfmt'
Plug 'godlygeek/tabular'
Plug 'hashivim/vim-packer'
Plug 'hashivim/vim-terraform'
Plug 'hashivim/vim-vagrant'
" Plug 'ianks/vim-tsx'
" Plug '/usr/local/opt/fzf'
Plug 'jvirtanen/vim-hcl'
Plug 'junegunn/fzf', { 'do': { -> fzf#install() }}
Plug 'junegunn/fzf.vim'
Plug 'kana/vim-fakeclip'
Plug 'kshenoy/vim-signature'
Plug 'leshill/vim-json'
" Plug 'leafgarland/typescript-vim'
Plug 'lilydjwg/colorizer'
Plug 'mattn/emmet-vim'
Plug 'morhetz/gruvbox'
" Plug 'neoclide/coc.nvim', {'branch': 'release'}
" Plug 'OmniSharp/omnisharp-vim'
" Plug 'pangloss/vim-javascript'
Plug 'pbrisbin/vim-mkdir'
" Plug 'leafOfTree/vim-vue-plugin'
Plug 'lilydjwg/colorizer'
Plug 'mattn/emmet-vim'
Plug 'morhetz/gruvbox'
Plug 'neoclide/coc.nvim', {'branch': 'release'}
" Plug 'OmniSharp/omnisharp-vim'
" Plug 'pangloss/vim-javascript'
Plug 'pbrisbin/vim-mkdir'
" Plug 'posva/vim-vue'
Plug 'PProvost/vim-ps1'
Plug 'Quramy/vim-js-pretty-template'
Plug 'Raimondi/delimitMate'
" Plug 'scrooloose/nerdcommenter'
Plug 'scrooloose/nerdtree'
Plug 'srcery-colors/srcery-vim'
" Plug 'StanAngeloff/php.vim'
Plug 'stephpy/vim-yaml'
" Plug 'storyn26383/vim-vue'
Plug 'tpope/vim-bundler'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-dadbod'
Plug 'tpope/vim-fugitive'
Plug 'tpope/vim-projectionist'
Plug 'tpope/vim-ragtag'
" Plug 'tpope/vim-rails'
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-surround'
Plug 'tpope/vim-bundler'
Plug 'uarun/vim-protobuf'
Plug 'vim-airline/vim-airline'
Plug 'vim-airline/vim-airline-themes'
Plug 'vim-ruby/vim-ruby'
Plug 'yggdroot/indentLine'
" if has('nvim')
"   Plug 'Shougo/denite.nvim', { 'do': ':UpdateRemotePlugins' }
" else
"   Plug 'Shougo/denite.nvim'
"   Plug 'roxma/nvim-yarp'
"   Plug 'roxma/vim-hug-neovim-rpc'
" endif
call plug#end()
" colo gruvbox
" colo srcery
syntax enable
" colorscheme srcery
syntax on

" General VIM settings
set hidden                                    " Allow buffer change w/o saving
set lazyredraw                                " Dont' update while executing macros
set expandtab                                 " Convert <tab> to spaces (2 or 4)
set history=1000
set relativenumber
set scrolloff=4
set shiftwidth=2                              "     then override with per filetype
set softtabstop=2                             "     specific settings via autocmd
set splitbelow
set splitright
set tabstop=2                                 " Two spaces per tab as default

" inoremap " ""<left>
" inoremap { {}<left>
" inoremap [ []<left>

" Leader commands
let mapleader = " "
map <leader>vi :tabe ~/.vimrc<cr>

nmap k gk
nmap j gj
nmap <leader>so :source $MYVIMRC<cr>
nmap <leader>ca gg"*yG

" Vtr stuff
let g:VtrUseVtrMaps = 1
let g:VtrOrientation = "h"

" YAML stuff
autocmd FileType yaml setlocal ts=2 sts=2 sw=2 expandtab

" ===== Plugin setup ===== "

" ===== Airline settings =====
" air-line
let g:airline_powerline_fonts = 1

" let g:airline_theme='tomorrow'
let g:airline_theme='srcery'

" let g:ale_linters = {
"       \ 'cs': ['OmniSharp'],
"       \ 'javascript': ['eslint'],
"       \ 'php': ['phpcs'],
"       \ 'ruby': ['rubocop'],
"       \ 'typescript' : ['tsserver'],
" \}
" let g:ale_fixers = {
"       \ 'javascript': ['prettier'],
"       \ 'typescript': ['prettier'],
"       \ 'ruby': ['rubocop'],
"       \ 'vue': ['eslint'],
"       \ 'scss': ['prettier'],
"       \ 'php': ['phpcbf']
" \}
" let g:ale_fix_on_save = 1
" let g:ale_linters_explicit = 1
" let g:airline#extensions#ale#enabled = 1

" ======= CoC Settings ========
" use <Tab> and <S-Tab> to navigate the trigger list
inoremap <expr> <Tab> pumvisible() ? "\<C-n>" : "\<Tab>"
inoremap <expr> <S-Tab> pumvisible() ? "\<C-p>" : "\<S-Tab>"

inoremap <expr> <cr> pumvisible() ? "\<C-y>" : "\<C-g>u\<CR>"

"=== Emmet settings ==="
let g:user_emmet_leader_key=','

"=== fzf settings ==="
" File finder
nmap <silent> <Leader>f :Files<CR>
nmap <silent> <Leader>F :GFiles<CR>

" Line Finder
nmap <silent> <Leader>l :BLines<CR>
nmap <silent> <Leader>L : Lines<CR>
nmap <silent> <Leader>' :Marks<CR>

" Other finders
" noremap <silent>

command! -bang -nargs=* Find call fzf#vim#grep('rg --column --line-number --no-heading --fixed-strings --ignore-case --no-ignore --hidden --follow --glob "!.git/*" --color "always" '.shellescape(<q-args>).'| tr -d "\017"', 1, <bang>0)

let g:fzf_action = {
  \ 'ctrl-t': 'tab split',
  \ 'ctrl-x': 'split',
  \ 'ctrl-v': 'vsplit' }

let g:fzf_layout = { 'down': '~40%' }

let g:fzf_colors =
\ { 'fg':      ['fg', 'Normal'],
  \ 'bg':      ['bg', 'Normal'],
  \ 'hl':      ['fg', 'Comment'],
  \ 'fg+':     ['fg', 'CursorLine', 'CursorColumn', 'Normal'],
  \ 'bg+':     ['bg', 'CursorLine', 'CursorColumn'],
  \ 'hl+':     ['fg', 'Statement'],
  \ 'info':    ['fg', 'PreProc'],
  \ 'border':  ['fg', 'Ignore'],
  \ 'prompt':  ['fg', 'Conditional'],
  \ 'pointer': ['fg', 'Exception'],
  \ 'marker':  ['fg', 'Keyword'],
  \ 'spinner': ['fg', 'Label'],
  \ 'header':  ['fg', 'Comment'] }

"=== Treat all svelte files as HTML ==="
au! BufNewFile,BufRead *.svelte set ft=html

" let g:NERDTreeWinPos = 0 
" let g:NERDTreeMinimalUI  = 1
" let g:NERDTreeShowHidden = 1

let g:javascript_plugin_flow = 1

let g:netrw_banner = 0
let g:netrw_browse_split = 2
let g:netrw_winsize = 25

let g:OmniSharp_server_stdio = 1

"=== vim-terraform settings === 
let g:terraform_align=1
let g:terraform_fmt_on_save=1

"=== indentLine plugin ===
let g:indentLine_char = '⦙'

" set formatoptions-=ro

autocmd FileType * setlocal formatoptions-=c formatoptions-=r formatoptions-=o
autocmd FileType typescript setlocal formatprg=prettier\ --parser\ typescript
autocmd StdinReadPre * let s:std_in=1
autocmd VimEnter * if argc() == 1 && isdirectory(argv()[0]) && !exists("s:std_in") | exe 'NERDTree' argv()[0] | wincmd p | ene | exe 'cd '.argv()[0] | endif

set number

""" Customize colors
func! s:my_colors_setup() abort
    " this is an example
    hi Pmenu guibg=#d7e5dc gui=NONE
    hi PmenuSel guibg=#b7c7b7 gui=NONE
    hi PmenuSbar guibg=#bcbcbc
    hi PmenuThumb guibg=#585858
endfunc

augroup colorscheme_coc_setup | au!
    au ColorScheme * call s:my_colors_setup()
augroup EN

" ======= color settings =======
let g:srcery_italic = 1
" colorscheme srcery
colorscheme srcery
" set guifont=IBM\ Plex\ Mono:h13
set guifont=Perplexed:h13
set guicursor=


" automatically rebalance windows on vim resize
autocmd VimResized * :wincmd =
" zoom a vim pane, <C-w>= to re-balance
nnoremap <leader>- :wincmd _<cr>:wincmd \|<cr>
nnoremap <leader>= :wincmd =<cr>

" FUNCTIONS
nmap =j :%!python3 -m json.tool<CR>
