package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type ResultType string

const (
	Boolean ResultType = "Boolean"
	String  ResultType = "String"
	Int     ResultType = "Int"
	Error   ResultType = "Error"
	Unknown ResultType = "Unknown"
)

type Kind string

const (
	Expression          Kind = "Expression"
	Atom                Kind = "Atom"
	DeclarationAtom     Kind = "DeclarationAtom"
	SymbolAtom          Kind = "SymbolAtom"
	ArgumentsExpression Kind = "ArgumentsExpression"
)

type Node struct {
	Kind          Kind
	Type          ResultType
	Value         []byte
	Nodes         []*Node
	ArgumentIndex int
}

type EvaluationResultValue any

type EvaluationResult struct {
	Type  ResultType
	Value EvaluationResultValue
	Error error
}

type Argument struct {
	Value any
	Type  ResultType
}

type StackFrame struct {
	Function  string
	Arguments []Argument
}

type FunctionEntry struct {
	Name      string
	Body      *Node
	Arguments *Node
}

type TestCase[T int | bool | string] struct {
	input  string
	output T
}

type TestResult struct {
	pass int
	fail int
}

var Red = "\033[31m"
var Green = "\033[32m"
var Reset = "\033[0m"

var array_int_test_cases = []TestCase[int]{
	{"1", 1},
	{"32", 32},
	{"255", 255},
	{"256", 256},
	{"65536", 65536},
	{"65537", 65537},
	{"16777216", 16777216},
	{"16777217", 16777217},
	{"(+ 1 2)", 3},
	{"(* 1 2)", 2},
	{"(* 5 3)", 15},
	{"(/ 5 5)", 1},
	{"(/ 6 3)", 2},
	{"(- 1 1)", 0},
	{"(- 1 2)", -1},
}

func evaluate_test[T int | bool | string](name string, array []TestCase[T]) TestResult {
	pass := 0
	fail := 0
	for index, element := range array {
		element_functions := []FunctionEntry{}
		results := []EvaluationResult{}
		nodes := []*Node{}
		subs := strings.Split(element.input, "\n")
		for _, sub_element := range subs {
			_, node, functions, _ := atom_node(sub_element, 0, []FunctionEntry{})
			nodes = append(nodes, node)
			element_functions = slices.Concat(element_functions, functions)
			r := evaluate_expression(node, element_functions)
			results = append(results, r)
			if r.Error != nil {
				fmt.Printf("\nFailed to evaluate due to %v", r.Error.Error())
			}
		}
		final := results[len(results)-1]

		if final.Value != element.output {
			fmt.Printf("\nFunctions declared: %v", element_functions)
			for i, r := range results {
				fmt.Printf("\nExpression %d: `%v`", index, element.input)
				fmt.Printf("\n%vFAIL%v", Red, Reset)
				fmt.Printf("\nEvaluated as %v: %v", r.Type, r.Value)
				print(nodes[i], "")
				fail += 1
			}

		} else {
			pass += 1
			fmt.Printf("\n%vPASS%v", Green, Reset)
		}
		fmt.Printf("\n---------------")
	}

	fmt.Printf("\n[%v] result: %v Pass %d | %v Fail %d | %v Total %d", name, Green, pass, Red, fail, Reset, pass+fail)
	return TestResult{pass, fail}
}

func main() {
	results := []TestResult{
		evaluate_test("int evaluation", []TestCase[int]{
			{"1", 1},
			{"32", 32},
			{"255", 255},
			{"256", 256},
			{"65536", 65536},
			{"65537", 65537},
			{"16777216", 16777216},
			{"16777217", 16777217},
			{"(+ 1 2)", 3},
			{"(* 1 2)", 2},
			{"(* 5 3)", 15},
			{"(/ 5 5)", 1},
			{"(/ 6 3)", 2},
			{"(- 1 1)", 0},
			{"(- 1 2)", -1},

			{"(defun doublen (n) (* n 2))\n (doublen 2)", 4},
		}),

		evaluate_test("bool evaluation", []TestCase[bool]{
			{"t", true},
			{"T", true},
			{"nil", false},
		}),

		evaluate_test("string evaluation", []TestCase[string]{
			{"\"Hello, Coding Challenges\"",
				"\"Hello, Coding Challenges\""},
		}),
	}

	pass := 0
	fail := 0

	for _, r := range results {
		pass += r.pass
		fail += r.fail
	}

	fmt.Printf("\n[full suite] result: %v Pass %d | %v Fail %d | %v Total %d", Green, pass, Red, fail, Reset, pass+fail)

	if fail > 0 {
		panic("there were failing tests")
	}

	array := []string{
		"()",
		":CC",
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
		functions := []FunctionEntry{}
		for _, sub_element := range strings.Split(element, "\n") {
			_, node, f, _ := atom_node(sub_element, 0, []FunctionEntry{})
			functions = slices.Concat(functions, f)
			fmt.Printf("\nExpression %d: `%v`", index, sub_element)
			r := evaluate_expression(node, functions)
			if r.Error != nil {
				fmt.Printf("\nFailed to evaluate due to %v", r.Error.Error())
			} else {
				fmt.Printf("\nEvaluated as %v %v", r.Type, r.Value)
				print(node, "")
			}
			fmt.Printf("\n---------------")
		}
	}
}

func print(node *Node, indent string) {
	value := ""
	if node.Kind == Atom && node.Type == Int {
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

func get_operator_bytes(node *Node) ([]byte, error) {
	if len(node.Nodes) < 1 /*|| node.Nodes[0].Kind != Atom*/ {
		return []byte{}, fmt.Errorf("unable to get operator atom %v", node.Value)
	}
	return node.Nodes[0].Value, nil
}

func binary_operator_int(node *Node, stack []StackFrame, functions []FunctionEntry, f func(int, int) int) EvaluationResult {
	a, e1 := operand_int(node, 1, stack, functions)
	b, e2 := operand_int(node, 2, stack, functions)
	if e1 != nil || e2 != nil {
		return EvaluationResult{Error, nil, first_error(e1, e2)}
	}
	return EvaluationResult{Int, f(a, b), nil}
}

func operand_int(node *Node, index int, stack []StackFrame, functions []FunctionEntry) (int, error) {
	r := evaluate(node.Nodes[index], stack, functions)
	if r.Type == Int {
		return r.Value.(int), r.Error
	}
	return 0, fmt.Errorf("Unable to evaluate operand from the node %v, value: %v", node, string(node.Value))
}

func first_error(input ...error) error {
	for _, item := range input {
		if item != nil {
			return item
		}
	}
	return nil
}

func evaluate_expression(node *Node, functions []FunctionEntry) EvaluationResult {
	return evaluate(node, []StackFrame{StackFrame{Function: "__entry_point", Arguments: []Argument{}}}, functions)
}

// node: symbol atom of call site
// first child is a function, followed by the arguments
func call_function(node *Node, stack []StackFrame, function FunctionEntry, functions []FunctionEntry) EvaluationResult {
	name := string(node.Nodes[0].Value)
	args := []Argument{}
	for i := 1; i < len(node.Nodes); i++ {
		r := evaluate(node.Nodes[i], stack, functions)
		if r.Error != nil {
			//todo: probably worth wrapping with more context
			return r
		}
		arg := Argument{r.Value, r.Type}
		args = append(args, arg)
	}
	// evaluate arguments
	if len(function.Arguments.Nodes) != len(args) {
		return EvaluationResult{Error: fmt.Errorf("arguments count provided to the function `%v` don't match declared function, expected: %d, actual: %d", name, len(function.Arguments.Nodes), len(args))}
	}
	for i := 0; i < len(args); i++ {
		if args[i].Type != function.Arguments.Nodes[i].Type {
			return EvaluationResult{Error: fmt.Errorf("the argument %v type provided to the function `%v` doesn't match declared function's argument type, expected: %v, actual: %v", string(function.Arguments.Nodes[i].Value), name, function.Arguments.Nodes[i].Type, args[i].Type)}
		}
	}
	// fmt.Printf("calling function %v args %v body %v", name, args, function.Body)
	frame := StackFrame{name, args}
	return evaluate(function.Body, append(stack, frame), functions)
}

func evaluate(node *Node, stack []StackFrame, functions []FunctionEntry) EvaluationResult {
	// int literal
	if node.Kind == Atom && node.Type == Int {
		return EvaluationResult{Int, bytes_to_int32(node.Value), nil}
	}
	// string literal
	if node.Kind == Atom && node.Type == String {
		return EvaluationResult{String, string(node.Value), nil}
	}
	// boolean literal
	if node.Kind == Atom && node.Type == Boolean {
		return EvaluationResult{Boolean, node.Value[0] == 1, nil}
	}

	// resolve variable
	if node.Kind == SymbolAtom && len(node.Nodes) == 0 && node.ArgumentIndex >= 0 && len(stack) > 0 && node.ArgumentIndex < len(stack[len(stack)-1].Arguments) {
		// todo: convert into a real stack
		arg_value := stack[len(stack)-1].Arguments[node.ArgumentIndex]
		return EvaluationResult{arg_value.Type, arg_value.Value, nil}
	}
	// call function
	if node.Kind == Expression && len(node.Nodes) > 0 && node.Nodes[0].Kind == Atom {
		for _, f := range functions {
			if f.Name == string(node.Nodes[0].Value) {
				return call_function(node, stack, f, functions)
			}
		}
	}

	// operators
	if node.Kind == Expression && len(node.Nodes) > 2 {
		oper, err := get_operator_bytes(node)
		if err != nil {
			return EvaluationResult{Error, nil, err}
		}
		if oper[0] == '+' {
			return binary_operator_int(node, stack, functions, func(a, b int) int {
				return a + b
			})
		}
		if oper[0] == '-' {
			return binary_operator_int(node, stack, functions, func(a, b int) int {
				return a - b
			})
		}
		if oper[0] == '*' {
			return binary_operator_int(node, stack, functions, func(a, b int) int {
				return a * b
			})
		}
		if oper[0] == '/' {
			return binary_operator_int(node, stack, functions, func(a, b int) int {
				return a / b
			})
		}
	}
	return EvaluationResult{Unknown, nil, nil}
}

func expression_node(input string, pos int, functions []FunctionEntry) (bool, *Node, []FunctionEntry, int) {
	for pos < len(input) && input[pos] == ' ' {
		pos += 1
	}
	if input[pos] == '(' {
		pos += 1
		nodes := []*Node{}
		types := []ResultType{}
		for pos < len(input) && input[pos] != ')' {
			// expression consists of atoms which can be other expressions, strings or basic atoms
			r, node, functions2, p := atom_node(input, pos, functions)
			// need to do it because atom_node returns functions and it can shadow existing variable
			functions = functions2
			if r {
				pos = p
				nodes = append(nodes, node)
				if node.Type != Unknown {
					types = append(types, node.Type)
				}
				// function declaration case
				if node.Kind == DeclarationAtom {
					// declaration := node
					// extract name
					r, node, functions, p = atom_node(input, pos, functions)
					node.Kind = SymbolAtom
					nodes = append(nodes, node)
					pos = p
					r2, arguments, body, p := extract_function(input, pos, functions)
					if r2 {
						// name symbol return type (function return type) corresponds body
						node.Type = body.Type
						nodes = append(nodes, arguments)
						nodes = append(nodes, body)
						pos = p
						functions = append(functions, FunctionEntry{Name: string(node.Value), Body: body, Arguments: arguments})
					}
				}
			}
		}

		if len(types) == 1 {
			// if the expression consists of operator and operands, if one of the operands has type
			// we can infer the type of the operand of unknown type
			if slices.Index([]byte("+-*/"), nodes[0].Value[0]) > -1 {
				// might want to setup nodes[0].Type = types[0]
				for i := 1; i < len(nodes); i++ {
					if nodes[i].Type == Unknown {
						nodes[i].Type = types[0]
					}
				}
			}
			return true, node(Expression, []byte{}, types[0], nodes), functions, pos + 1
		}
		return true, node(Expression, []byte{}, Unknown, nodes), functions, pos + 1
	}
	return false, nil, functions, pos
}

func extract_function(input string, pos int, functions []FunctionEntry) (bool, *Node, *Node, int) {
	// extract arguments
	r, node, functions, p := expression_node(input, pos, functions)
	if r {
		node.Kind = ArgumentsExpression
		arguments := node
		args := []string{}
		for _, arg := range node.Nodes {
			arg.Kind = SymbolAtom
			args = append(args, string(arg.Value))
		}

		// extract body
		r, node, functions, p = expression_node(input, p, functions)
		body := node
		for _, n := range node.Nodes {
			reference := slices.Index(args, string(n.Value))
			if reference > -1 {
				n.Kind = SymbolAtom
				n.ArgumentIndex = reference
				arguments.Nodes[reference].Type = n.Type
			}
		}
		return true, arguments, body, p
	} else {
		return false, nil, nil, -1
	}
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

func decl_node(input string) (bool, *Node) {
	if input == "defun" {
		return true, node(DeclarationAtom, []byte(input), Unknown, []*Node{})
	}
	return false, nil
}

func atom_node(input string, pos int, functions []FunctionEntry) (bool, *Node, []FunctionEntry, int) {
	// atoms are split by spaces
	for pos < len(input) && input[pos] == ' ' {
		pos += 1
	}
	// atom can be an expression consisting of other atoms/expressions
	r, n, functions, p := expression_node(input, pos, functions)
	if r {
		return r, n, functions, p
	}
	// atom can be a string atom
	r, n, p = string_node(input, pos)
	if r {
		return r, n, functions, p
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
		return true, n, functions, pos
	}
	r, n = bool_node(val)
	if r {
		return true, n, functions, pos
	}
	r, n = decl_node(val)
	if r {
		return true, n, functions, pos
	}

	return true, node(Atom, []byte(input[begin:pos]), Unknown, []*Node{}), functions, pos //keep separator in the buffer
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
