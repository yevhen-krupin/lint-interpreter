go -C ../core test \
&& go -C ../core build  \
&& go build \
&& go version -m ./lisp-repl-server | grep 'vcs.revision' | cut -d "=" -f2 | cut -b 1-8 | xargs printf "v0.1.%s" > ./static/version.txt 