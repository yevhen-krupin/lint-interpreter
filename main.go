package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

func main() {
	buf := bufio.NewReader(os.Stdin)
	st := SymbolTable{Functions: &map[string]FunctionEntry{}}
	runtime := NewRuntime(&st)
	lisp := Lisp()
	if len(os.Args) > 2 {
		//-i input
		//-o output file
		input, oki := argument("-i", "--input")
		output, oko := argument("-o", "--output")
		if oko {
			defer to_disk(output, "log.log").Close()
		}
		if oki {
			expressions := strings.Split(input, "\\n")
			log.Println("input provided", expressions)
			for _, ex := range expressions {
				ast, _ := ParseAst(strings.TrimLeft(ex, "\n"), &st, lisp)
				r := runtime.EvaluateAst(ast)
				fmt.Println("Evaluated", r.Type, r.Value, r.Error)
			}
			return
		}
	}
	for true {
		fmt.Print("> ")
		sentence, err := buf.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
		} else {
			ast, _ := ParseAst(string(sentence[:len(sentence)-1]), &st, lisp)
			print(ast, "")
			r := runtime.evaluate(ast)
			fmt.Println("Evaluated", r.Type, r.Value, r.Error)
		}
	}

}

func argument(name string, alias string) (string, bool) {
	// 0 - executable name
	i := slices.Index(os.Args, name)
	if i < 1 {
		i = slices.Index(os.Args, alias)
	}
	// found, next entry should be the value
	if i >= 1 {
		if i > len(os.Args)-1 || os.Args[i+1][0] == '-' {
			log.Println("the argument", name, "or", alias, "was expected but not provided")
			return "", false
		}
		return os.Args[i+1], true
	}
	return "", false
}
