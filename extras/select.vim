" These enable visual mode selection from insert mode with Shift+Left/Right,
" and let you select your transcription using the "select" phrase.

nnoremap <S-Left> v
inoremap <S-Left> <Esc>v
vnoremap <S-Left> h
nnoremap <S-Right> v
inoremap <S-Right> <Esc>lv
vnoremap <S-Right> l

" These let you format your selection.
" For example, "minus cap" makes your selection camelCase.

" a, kind, of, list, case
vnoremap <silent> -, :s/\%V /, /g<CR>
" kebab-case
vnoremap <silent> -- :s/\%V /-/g<CR>
" dot.case
vnoremap <silent> -. :s/\%V /./g<CR>
" colon::case
vnoremap <silent> -: :s/\%V /::/g<CR>
" snake_case
vnoremap <silent> -_ :s/\%V /_/g<CR>
" UPPER_SNAKE_CASE
vnoremap <silent> -u :s/\%V./\u&/g <bar> s/\%V /_/g<CR>
" camelCase
vnoremap <silent> -c :s/\%V \?\<\(.\)/\u\1/g<CR>
" MixedCase
vnoremap <silent> -m :s/\%V.\zs \?\<\(.\)/\u\1/g<CR>
" Title Case
vnoremap <silent> -t :s/\%V\<./\u&/g<CR>
" allsmashedtogether
vnoremap <silent> -<space> :s/\%V //g<CR>
