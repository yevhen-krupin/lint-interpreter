package main

import (
	"fmt"
	"strconv"
)

type ResultType string

const (
	Boolean ResultType = "Boolean"
	String  ResultType = "String"
	Int     ResultType = "Int"
	Unknown ResultType = "Unknown"
)

type Kind string

const (
	Expression Kind = "Expression"
	Atom       Kind = "Atom"
)

type Node struct {
	Kind  Kind
	Type  ResultType
	Value []byte
	Nodes []*Node
}

func main() {
	array := []string{
		"()",
		"1",
		"32",
		"255",
		"256",
		"65536",
		"65537",
		"16777216",
		"16777217",
		"t",
		"\"Hello, Coding Challenges\"",
		":CC",
		"(+ 1 2)",
		"(* 1 2)",
		"(* 5 3)",
		"(defun doublen (n) (* n 2))",
		"(defun hello() \"Hello Coding Challenges\")",
		"(format t \"Hello, Coding Challenge World World\")",
		"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1))" +
			"      (fib (- n 2)))))",
	}

	for index, element := range array {
		_, node, _ := atom_node(element, 0)
		fmt.Printf("\nExpression %d: `%v`", index, element)
		t, b, i, s := evaluate(node)
		fmt.Printf("\nEvaluated as %v %t %d `%v`", t, b, i, s)
		print(node, "")
		fmt.Printf("\n---------------")
	}
}

func print(node *Node, indent string) {
	value := ""
	if node.Type == Int {
		value = strconv.Itoa(bytes_to_int32(node.Value))
	}
	if node.Type == String || node.Type == Unknown {
		value = string(node.Value)
	}
	fmt.Printf("\n%v- %v: %v: %v [%v]", indent, node.Kind, node.Type, node.Value, value)
	for _, n := range node.Nodes {
		print(n, indent+"  ")
	}
}

func evaluate(node *Node) (ResultType, bool, int, string) {
	if node.Type == Int {
		return Int, false, bytes_to_int32(node.Value), ""
	}
	if node.Type == String {
		return String, false, 0, string(node.Value)
	}
	if node.Type == Boolean {
		return Boolean, node.Value[0] == 1, 0, ""
	}
	if node.Kind == Expression && len(node.Nodes) > 0 {
		if node.Nodes[0].Kind == Atom && node.Nodes[0].Value[0] == '+' {
			i := 1
			r := 0
			for i < len(node.Nodes) {
				r += bytes_to_int32(node.Nodes[i].Value)
				i += 1
			}
			return Int, false, r, ""
		}

		if node.Nodes[0].Kind == Atom && node.Nodes[0].Value[0] == '*' {
			i := 1
			r := 1
			for i < len(node.Nodes) {
				r *= bytes_to_int32(node.Nodes[i].Value)
				i += 1
			}
			return Int, false, r, ""
		}
	}
	return Unknown, false, 0, ""
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
		return true, node(Expression, []byte{}, Unknown, nodes), pos + 1
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
		return true, node(Atom, []byte(input[begin:pos]), String, []*Node{}), pos + 1
	}
	return false, nil, pos
}

func int_node(input string) (bool, *Node) {
	i, err := strconv.Atoi(input)
	if err == nil {
		return true, node(Atom, *int32_to_bytes(i), Int, []*Node{})
	}
	return false, nil
}

func bool_node(input string) (bool, *Node) {
	if input == "t" || input == "T" {
		return true, node(Atom, []byte{1}, Boolean, []*Node{})
	}
	if input == "nil" {
		return true, node(Atom, []byte{0}, Boolean, []*Node{})
	}
	return false, nil
}

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

	val := input[begin:pos]
	r, n = int_node(val)
	if r {
		return true, n, pos
	}
	r, n = bool_node(val)
	if r {
		return true, n, pos
	}
	return true, node(Atom, []byte(input[begin:pos]), Unknown, []*Node{}), pos //keep separator in the buffer
}

func node(kind Kind, value []byte, rt ResultType, nodes []*Node) *Node {
	var node Node
	node.Type = rt
	node.Kind = kind
	node.Value = value
	node.Nodes = nodes
	return &node
}

func int32_to_bytes(i int) *[]byte {
	buf := make([]byte, 4)
	buf[3] = byte(i & 0xFF)
	buf[2] = byte((i & 0xFF00) >> 8)
	buf[1] = byte((i & 0xFF0000) >> 16)
	buf[0] = byte((i & 0xFF000000) >> 24)
	return &buf
}

func bytes_to_int32(bytes []byte) int {
	v := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	return v //bytes[0]*256*256*256 + bytes[1]*256*256 | bytes[2]*256 | bytes[3]
}
