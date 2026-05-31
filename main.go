package main

import (
	"fmt"
)

type Node struct {
	kind  string
	value string
	nodes []*Node
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
	}
}

func print(node *Node, indent int) {
	fmt.Printf("\n- %v: %v", node.kind, node.value)
	for _, n := range node.nodes {
		print(n, indent+2)
	}
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
