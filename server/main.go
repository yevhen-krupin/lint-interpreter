package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	interpreter "ykrupin.com/lisp-interpreter-core"
)

var lisp = interpreter.Lisp()

func main() {
	bind, okb := interpreter.GetArgument("-b", "--bind")
	if okb {
		http.HandleFunc("/repl", serve)
		log.Printf("serve /repl as ws://%v/repl", bind)
		http.Handle("/", http.FileServer(http.Dir("./static")))
		log.Printf("serve static home / as https://%v/", bind)
		log.Println("binding address", bind)
		log.Fatal(http.ListenAndServe(bind, nil))
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	log.Println("trigger")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	runtime := interpreter.NewRuntime()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		expressions := strings.Split(string(message), "\\n")
		log.Println("input provided", expressions)
		for _, ex := range expressions {
			ast, _ := interpreter.ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
			r := runtime.EvaluateAst(ast)
			err = c.WriteMessage(mt, []byte(fmt.Sprintln("Evaluated", r.Type, r.Value, r.Error)))
		}
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
