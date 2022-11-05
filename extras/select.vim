" These enable visual mode selection from insert mode with Shift+Left/Right,
" and let you select your transcription using the "select" phrase.

nnoremap <S-Left> v
inoremap <S-Left> <Esc>v
vnoremap <S-Left> h
nnoremap <S-Right> v
inoremap <S-Right> <Esc>lv
vnoremap <S-Right> l

" These let you format your selection.
" For example, "score cap" makes your selection camelCase.

" kebab-case
vnoremap <silent> _- :s/\%V /-/g<CR>`>a
" dot.case
vnoremap <silent> _. :s/\%V /./g<CR>`>a
" colon::case
vnoremap <silent> _: <Esc>`>a <Esc>:s/\%V /::/g<CR>`<f s
" snake_case
vnoremap <silent> __ :s/\%V /_/g<CR>`>a
" UPPER_SNAKE_CASE
vnoremap <silent> _u :s/\%V./\u&/g <bar> s/\%V /_/g<CR>`>a
" camelCase
vnoremap <silent> _c <Esc>`>a <Esc>:s/\%V \<\(.\)/\u\1/g<CR>`<f s
" MixedCase
vnoremap <silent> _m <Esc>`>a <Esc>:s/\%V \?\<\(.\)/\u\1/g<CR>`<f s
" Title Case
vnoremap <silent> _t :s/\%V\<./\u&/g<CR>`>a
" allsmashedtogether
vnoremap <silent> _<space> <Esc>`>a <Esc>:s/\%V //g<CR>`<f s
" a, kind, of, list, case
vnoremap <silent> _, <Esc>`>a  <Esc>:s/\%V /, /g<CR>`>/  <CR>2s
" "a", "kind", "of", "quoted", "list", "case"
vnoremap <silent> _" <Esc>`>a  <Esc>:s/\%V /", "/g<CR>`<i"<Esc>`>/  <CR>2s"
" 'a', 'kind', 'of', 'quoted', 'list', 'case'
vnoremap <silent> _' <Esc>`>a  <Esc>:s/\%V /', '/g<CR>`<i'<Esc>`>/  <CR>2s'
