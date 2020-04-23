call plug#begin('~/.vim/plugged')
Plug 'airblade/vim-gitgutter'
Plug 'Akulen/vim-dcrpc'
Plug 'christoomey/vim-tmux-navigator'
Plug 'christoomey/vim-tmux-runner'
Plug 'dense-analysis/ale'
Plug 'dikiaap/minimalist'
Plug 'fatih/vim-go'
Plug 'ianks/vim-tsx'
Plug '/usr/local/opt/fzf'
Plug 'junegunn/fzf.vim'
Plug 'leshill/vim-json'
Plug 'leafgarland/typescript-vim'
Plug 'leafOfTree/vim-vue-plugin'
Plug 'mattn/emmet-vim'
Plug 'morhetz/gruvbox'
Plug 'neoclide/coc.nvim', {'branch': 'release'}
Plug 'OmniSharp/omnisharp-vim'
Plug 'pangloss/vim-javascript'
Plug 'pbrisbin/vim-mkdir'
Plug 'posva/vim-vue'
Plug 'Quramy/vim-js-pretty-template'
Plug 'scrooloose/nerdcommenter'
Plug 'scrooloose/nerdtree'
Plug 'srcery-colors/srcery-vim'
Plug 'stephpy/vim-yaml'
Plug 'tpope/vim-bundler'
Plug 'tpope/vim-dadbod'
Plug 'tpope/vim-fugitive'
Plug 'tpope/vim-projectionist'
Plug 'tpope/vim-ragtag'
Plug 'tpope/vim-rails'
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-surround'
Plug 'tpope/vim-bundler'
Plug 'vim-airline/vim-airline'
Plug 'vim-airline/vim-airline-themes'
Plug 'vim-ruby/vim-ruby'
" if has('nvim')
"   Plug 'Shougo/denite.nvim', { 'do': ':UpdateRemotePlugins' }
" else
"   Plug 'Shougo/denite.nvim'
"   Plug 'roxma/nvim-yarp'
"   Plug 'roxma/vim-hug-neovim-rpc'
" endif
call plug#end()
" colo gruvbox
colo srcery
syntax on

let mapleader = " "

" Vtr stuff
let g:VtrUseVtrMaps = 1
let g:VtrOrientation = "h"

" ===== Plugin setup ===== "

" ===== Airline settings =====
" air-line
let g:airline_powerline_fonts = 1

" let g:airline_theme='tomorrow'
let g:airline_theme='srcery'

let g:ale_linters = {
      \ 'cs': ['OmniSharp'],
      \ 'javascript': ['eslint'],
      \ 'ruby': ['rubocop'],
      \ 'typescript' : ['tsserver', 'tslint'],
\}
let g:ale_fixers = {
      \ 'javascript': ['eslint'],
      \ 'typescript': ['prettier'],
      \ 'ruby': ['rubocop'],
      \ 'vue': ['eslint'],
      \ 'scss': ['prettier'],
      \ 'html': ['prettier']
\}
let g:ale_fix_on_save = 1
let g:ale_linters_explicit = 1
let g:airline#extensions#ale#enabled = 1

" ======= CoC Settings ========
" use <Tab> and <S-Tab> to navigate the trigger list
inoremap <expr> <Tab> pumvisible() ? "\<C-n>" : "\<Tab>"
inoremap <expr> <S-Tab> pumvisible() ? "\<C-p>" : "\<S-Tab>"

inoremap <expr> <cr> pumvisible() ? "\<C-y>" : "\<C-g>u\<CR>"


" ======= Denite Settings ========
" try
" call denite#custom#var('file/rec', 'command', ['rg', '--files', '--glob', '!.git'])
" 
" " Use ripgrep in place of "grep"
" call denite#custom#var('grep', 'command', ['rg'])
" 
" " Custom options for ripgrep
" "   --vimgrep:  Show results with every match on it's own line
" "   --hidden:   Search hidden directories and files
" "   --heading:  Show the file name above clusters of matches from each file
" "   --S:        Search case insensitively if the pattern is all lowercase
" call denite#custom#var('grep', 'default_opts', ['--hidden', '--vimgrep', '--heading', '-S'])
" 
" " Recommended defaults for ripgrep via Denite docs
" call denite#custom#var('grep', 'recursive_opts', [])
" call denite#custom#var('grep', 'pattern_opt', ['--regexp'])
" call denite#custom#var('grep', 'separator', ['--'])
" call denite#custom#var('grep', 'final_opts', [])
" 
" " Remove date from buffer list
" call denite#custom#var('buffer', 'date_format', '')
" 
" " Custom options for Denite
" "   auto_resize             - Auto resize the Denite window height automatically.
" "   prompt                  - Customize denite prompt
" "   direction               - Specify Denite window direction as directly below current pane
" "   winminheight            - Specify min height for Denite window
" "   highlight_mode_insert   - Specify h1-CursorLine in insert mode
" "   prompt_highlight        - Specify color of prompt
" "   highlight_matched_char  - Matched characters highlight
" "   highlight_matched_range - matched range highlight
" let s:denite_options = {'default' : {
" \ 'split': 'floating',
" \ 'start_filter': 1,
" \ 'auto_resize': 1,
" \ 'source_names': 'short',
" \ 'prompt': 'λ ',
" \ 'highlight_matched_char': 'QuickFixLine',
" \ 'highlight_matched_range': 'Visual',
" \ 'highlight_window_background': 'Visual',
" \ 'highlight_filter_background': 'DiffAdd',
" \ 'winrow': 1,
" \ 'vertical_preview': 1
" \ }}
" 
" " Loop through denite options and enable them
" function! s:profile(opts) abort
"   for l:fname in keys(a:opts)
"     for l:dopt in keys(a:opts[l:fname])
"       call denite#custom#option(l:fname, l:dopt, a:opts[l:fname][l:dopt])
"     endfor
"   endfor
" endfunction
" 
" call s:profile(s:denite_options)
" catch
"   echo 'Denite not installed. It should work after running :PlugInstall'
" endtry

"=== Denite shortcuts ==="
" nmap ; :Denite buffer<CR>
" nmap <leader>t :DeniteProjectDir file/rec<CR>
" nnoremap <leader>g :<C-u>Denite grep:. -no-empty<CR>
" nnoremap <leader>j :<C-u>DeniteCursorWord grep:.<CR>

"=== Emmet settings ==="
let g:user_emmet_leader_key=','

"=== fzf settings ==="
" File finder
nmap <Leader>f :Files<CR>
nmap <Leader>F :GFiles<CR>

" Line Finder
nmap <Leader>l :BLines<CR>
nmap <Leader>L : Lines<CR>
nmap <Leader>' :Marks<CR>


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
let g:NERDTreeMinimalUI  = 1
let g:NERDTreeShowHidden = 1

let g:javascript_plugin_flow = 1

let g:netrw_banner = 0
let g:netrw_browse_split = 2
let g:netrw_winsize = 25

let g:OmniSharp_server_stdio = 1

" set formatoptions-=ro
set tabstop=2 shiftwidth=2 expandtab
set splitbelow
set splitright

autocmd FileType * setlocal formatoptions-=c formatoptions-=r formatoptions-=o
autocmd FileType typescript setlocal formatprg=prettier\ --parser\ typescript
autocmd StdinReadPre * let s:std_in=1
autocmd VimEnter * if argc() == 1 && isdirectory(argv()[0]) && !exists("s:std_in") | exe 'NERDTree' argv()[0] | wincmd p | ene | exe 'cd '.argv()[0] | endif

set number

" ======= srcery settings =======
let g:srcery_italic = 1
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
nmap =j :%!python -m json.tool<CR>
