package main

import (
	"fmt"
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
		results := []EvaluationResult{}
		nodes := []*Node{}
		subs := strings.SplitSeq(element.input, "\n")
		runtime := NewRuntime()
		for sub_element := range subs {
			ast, _ := ParseAst(sub_element)
			nodes = append(nodes, ast.Root)
			r := runtime.EvaluateAst(ast)
			results = append(results, r)
			if r.Error != nil {
				fmt.Printf("\nFailed to evaluate due to %v", r.Error.Error())
			}
		}
		final := results[len(results)-1]

		if final.Value != element.output {
			fmt.Printf("\nFunctions declared: %v", runtime.Functions)
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
		runtime := NewRuntime()
		for sub_element := range strings.SplitSeq(element, "\n") {
			ast, _ := ParseAst(sub_element)
			fmt.Printf("\nExpression %d: `%v`", index, sub_element)
			r := runtime.EvaluateAst(ast)
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
	if node.Type == String || node.Type == Unknown || node.Kind == ArgumentDeclaration || node.Kind == ArgumentVariable {
		value = string(node.Value)
	}
	fmt.Printf("\n%v- %v: %v: %v [%v]", indent, node.Kind, node.Type, node.Value, value)
	for _, n := range node.Nodes {
		print(n, indent+"  ")
	}
}

func coalesce[TIn any, TOut comparable](in TIn, nodes ...func(TIn) TOut) TOut {
	var zero TOut
	for _, f := range nodes {
		n := f(in)
		if n != zero {
			return n
		}
	}
	return zero
}
