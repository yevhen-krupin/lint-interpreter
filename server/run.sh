go -C ../core test \
&& go  -C ../core build  \
&& go build \
&& ./lisp-repl-server -o build/test/work --bind 127.0.0.1:8080