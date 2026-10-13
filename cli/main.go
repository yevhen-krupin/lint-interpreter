package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	interpreter "ykrupin.com/lisp-interpreter-core"
)

var lisp = interpreter.Lisp()

func main() {
	//-e input expression
	//-i input file
	//-o output file
	expresion, oke := interpreter.GetArgument("-e", "--expression")
	input, oki := interpreter.GetArgument("-i", "--input")
	output, oko := interpreter.GetArgument("-o", "--output")
	if oke && oki {
		log.Fatalln("Error: Only one of two arguments should be supplied in the same time: -e/--expression or -i/--input")
	}
	if oko {
		defer interpreter.ToDisk(output, "log.log").Close()
	}
	if oki {
		file(input)
	} else if oke {
		oneshot(expresion)
	} else {
		repl()
	}
}

func oneshot(expresion string) {
	runtime := interpreter.NewRuntime()
	expressions := strings.Split(expresion, "\\n")
	log.Println("expression provided", expressions)
	for _, ex := range expressions {
		ast, _ := interpreter.ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
		r := runtime.EvaluateAst(ast)
		fmt.Println("Evaluated", r.Type, r.Value, r.Error)
	}
}

func file(input string) {
	if b, err := os.ReadFile(input); err == nil {
		runtime := interpreter.NewRuntime()
		str := string(b)
		expressions := strings.Split(str, "\n")
		log.Println("file input provided", input, "content:", strings.Join(expressions, ","))
		for _, ex := range expressions {
			ast, _ := interpreter.ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
			r := runtime.EvaluateAst(ast)
			fmt.Println("Evaluated", r.Type, r.Value, r.Error)
		}
	} else {
		log.Fatal(err)
	}
}

func repl() {
	runtime := interpreter.NewRuntime()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sentence, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
		} else {
			ast, _ := interpreter.ParseAst(string(sentence[:len(sentence)-1]), runtime.SymbolTable, lisp)
			print(ast, "")
			r := runtime.EvaluateAst(ast)
			fmt.Println("Evaluated", r.Type, r.Value, r.Error)
		}
	}
}
