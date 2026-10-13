module ykrupin.com/lisp-repl-server

go 1.26.4

replace ykrupin.com/lisp-interpreter-core => ../core

require (
	github.com/gorilla/websocket v1.5.3
	ykrupin.com/lisp-interpreter-core v0.0.0-00010101000000-000000000000
)
