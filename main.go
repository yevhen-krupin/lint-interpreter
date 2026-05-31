package main

import (
	"fmt"
)

type TokenDefinition struct {
	open  byte
	close byte
	kind  string
}

type Token struct {
	kind       string
	begin      int
	end        int
	valueBytes []byte
}

type Node struct {
	kind  string
	value string
	nodes []*Node
}

type Expression struct {
	terms []string
}

func main() {
	array := []string{
		"()",
		"1",
		"t",
		"\"Hello, Coding Challenges\"",
		":CC",
		"(+ 1 2)",
		"(defun hello() \"Hello Coding Challenges\")",
		"(format t \"Hello, Coding Challenge World World\")",
	}

	for index, element := range array {
		_, node, _ := atom_node(element, 0)
		fmt.Printf("\nExpression %d", index)
		print(node, 0)
		// result, err := tokenize(element)
		// if err != nil {
		// fmt.Printf("\n[%d] error: %s", index, err.Error())
		// } else {
		// fmt.Printf("\n[%d] type: %s start: %d end: %d", index, result[0].kind, result[0].begin, result[0].end)
		// }
	}
}

func print(node *Node, indent int) {
	fmt.Printf("\n- %v: %v", node.kind, node.value)
	for _, n := range node.nodes {
		print(n, indent+2)
	}
}

func tokenize(input string) ([]*Token, error) {
	var result []*Token
	for pos := 0; pos < len(input); pos++ {
		// while true here through the list of type definitions
		success, token, i := process(input, pos)
		if success {
			result = append(result, token)
			pos = i
		}
	}
	return result, nil
}

func process(input string, pos int) (bool, *Token, int) {
	definitions := []TokenDefinition{
		{
			kind:  "s-expression",
			open:  '(',
			close: ')',
		},
		{
			kind:  "string-atom",
			open:  '"',
			close: '"',
		},
	}
	for _, definition := range definitions {
		b, token, i := extract(input, definition, pos)
		if b == true {
			return true, token, i
		}

		b, token, i = atom(input, pos)
		if b == true {
			return true, token, i
		}
	}
	return false, nil, -1
}

func expression_node(input string, pos int) (bool, *Node, int) {
	if input[pos] == '(' {
		pos += 1
		nodes := []*Node{}
		for pos < len(input) && input[pos] != ')' {
			// expression consists of atoms which can be other expressions, strings or basic atoms
			r, node, p := atom_node(input, pos)
			if r {
				pos = p
				nodes = append(nodes, node)
			}
		}
		return true, node("expression", "", nodes), pos + 1
	}
	return false, nil, pos
}

func string_node(input string, pos int) (bool, *Node, int) {
	if input[pos] == '"' {
		begin := pos
		// include open quote
		pos += 1
		for pos < len(input) && input[pos] != '"' {
			pos += 1
		}
		// include closing quote
		pos += 1
		return true, node("string", input[begin:pos], []*Node{}), pos + 1
	}
	return false, nil, pos
}

// "(defun hello() \"Hello Coding Challenges\")",
func atom_node(input string, pos int) (bool, *Node, int) {
	// atoms are split by spaces
	for pos < len(input) && input[pos] == ' ' {
		pos += 1
	}
	// atom can be an expression consisting of other atoms/expressions
	r, n, p := expression_node(input, pos)
	if r {
		return r, n, p
	}
	// atom can be a string atom
	r, n, p = string_node(input, pos)
	if r {
		return r, n, p
	}
	// can handle more corner cases here

	// extract atom
	begin := pos
	// get all while it is not a separator or an end of current expression or a begin of a new one
	for pos < len(input) && input[pos] != ' ' && input[pos] != ')' && input[pos] != '(' {
		pos += 1
	}
	return true, node("atom", input[begin:pos], []*Node{}), pos //keep separator in the buffer
}

func node(kind string, value string, nodes []*Node) *Node {
	var node Node
	node.kind = kind
	node.value = value
	node.nodes = nodes
	return &node
}

func token(kind string, begin int, end int) *Token {
	var tstr Token
	tstr.kind = kind
	tstr.begin = begin
	tstr.end = end
	return &tstr
}

func atom(input string, pos int) (bool, *Token, int) {
	begin := pos
	// fmt.Printf("\n fallback to atom at pos %d : %v", pos, input[pos])
	for pos < len(input)+1 {
		if pos == len(input) || input[pos] == ' ' {
			// fmt.Printf("\n atom ends at pos %d", pos-1)
			return true, token("atom", begin, pos-1), pos + 1
		}
		pos += 1
	}
	return false, nil, pos
}

func extract(input string, definition TokenDefinition, pos int) (bool, *Token, int) {
	// fmt.Printf("\n definition %v expects %v, there is %v", definition.kind, definition.open, input[pos])
	if input[pos] == definition.open {
		begin := pos
		pos += 1
		// fmt.Printf("\n detected open %v of %v at pos %d", definition.open, definition.kind, pos)
		for pos < len(input) {
			if input[pos] == definition.close {
				// fmt.Printf("\n detected close %v of %v at pos %d", definition.close, definition.kind, pos)
				return true, token(definition.kind, begin, pos), pos + 1
			}
			pos += 1
		}
	}
	return false, &Token{}, pos
}
