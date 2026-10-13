package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var lisp = Lisp()

func main() {
	//-e input expression
	//-i input file
	//-o output file
	expresion, oke := GetArgument("-e", "--expression")
	input, oki := GetArgument("-i", "--input")
	output, oko := GetArgument("-o", "--output")
	if oke && oki {
		log.Fatalln("Error: Only one of two arguments should be supplied in the same time: -e/--expression or -i/--input")
	}
	if oko {
		defer to_disk(output, "log.log").Close()
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
	runtime := runtime()
	expressions := strings.Split(expresion, "\\n")
	log.Println("expression provided", expressions)
	for _, ex := range expressions {
		ast, _ := ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
		r := runtime.EvaluateAst(ast)
		fmt.Println("Evaluated", r.Type, r.Value, r.Error)
	}
}

func file(input string) {
	if b, err := os.ReadFile(input); err == nil {
		runtime := runtime()
		str := string(b)
		expressions := strings.Split(str, "\n")
		log.Println("file input provided", input, "content:", strings.Join(expressions, ","))
		for _, ex := range expressions {
			ast, _ := ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
			r := runtime.EvaluateAst(ast)
			fmt.Println("Evaluated", r.Type, r.Value, r.Error)
		}
	} else {
		log.Fatal(err)
	}
}

func repl() {
	runtime := runtime()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sentence, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
		} else {
			ast, _ := ParseAst(string(sentence[:len(sentence)-1]), runtime.SymbolTable, lisp)
			print(ast, "")
			r := runtime.evaluate(ast)
			fmt.Println("Evaluated", r.Type, r.Value, r.Error)
		}
	}
}

func runtime() Runtime {
	return NewRuntime(&SymbolTable{Functions: &map[string]FunctionEntry{}})
}
