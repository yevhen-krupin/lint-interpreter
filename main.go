package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

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

func evaluate_test[T int | bool | string](name string, array []TestCase[T]) TestResult {
	pass := 0
	fail := 0
	for index, element := range array {
		element_functions := []FunctionEntry{}
		results := []EvaluationResult{}
		nodes := []*Node{}
		subs := strings.SplitSeq(element.input, "\n")
		for sub_element := range subs {
			ast, _ := ParseAst(sub_element)
			nodes = append(nodes, ast.Root)
			element_functions = slices.Concat(element_functions, *ast.Functions)
			r := evaluate_expression(ast.Root, element_functions)
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
			}
			fail += 1

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
			{"(defun hello() (\"Hello Coding Challenges\")\n(hello)",
				"\"Hello Coding Challenges\""},
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
		"(format t \"Hello, Coding Challenge World World\")",
		"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1))" +
			"      (fib (- n 2)))))",
	}

	for index, element := range array {
		functions := []FunctionEntry{}
		for sub_element := range strings.SplitSeq(element, "\n") {
			ast, _ := ParseAst(sub_element)
			functions = slices.Concat(functions, *ast.Functions)
			fmt.Printf("\nExpression %d: `%v`", index, sub_element)
			r := evaluate_expression(ast.Root, functions)
			if r.Error != nil {
				fmt.Printf("\nFailed to evaluate due to %v", r.Error.Error())
			} else {
				fmt.Printf("\nEvaluated as %v %v", r.Type, r.Value)
				print(ast.Root, "")
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
	return evaluate(node, []StackFrame{{Function: "__entry_point", Arguments: []Argument{}}}, functions)
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

	//fmt.Printf("calling function %v args %v body %v", name, args, function.Body)
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
	// expression that just return something
	if node.Kind == Expression && len(node.Nodes) == 1 {
		return evaluate(node.Nodes[0], stack, functions)
	}

	return EvaluationResult{Unknown, nil, nil}
}
