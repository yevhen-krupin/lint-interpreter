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

type TestCase[T int | bool | string | []*Token] struct {
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

func tokenizer_test(array []TestCase[[]*Token]) TestResult {
	pass := 0
	fail := 0
	for index, element := range array {
		tokens := slices.Collect(tokenize(element.input))
		if !eq(tokens, element.output) {
			fmt.Printf("\nInput: %d: `%v`", index, element.input)
			fmt.Printf("\n%vFAIL%v", Red, Reset)
			fmt.Printf("\nTokenized as %v, expected %v", TokenStream(tokens), TokenStream(element.output))
			fail += 1
		} else {
			pass += 1
			fmt.Printf("\n%vPASS%v", Green, Reset)
		}
		fmt.Printf("\n---------------")
	}

	fmt.Printf("\n[%v] result: %v Pass %d | %v Fail %d | %v Total %d", "tokenizer", Green, pass, Red, fail, Reset, pass+fail)
	return TestResult{pass, fail}
}

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

			fmt.Printf("\nInput: %d: `%v`", index, element.input)
			fmt.Printf("\n%vFAIL%v", Red, Reset)
			fmt.Printf("\nFunctions declared: %v", runtime.Functions)
			for i, r := range results {
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
		tokenizer_test(
			[]TestCase[[]*Token]{
				{"1", t(d(1))},
				{"32", t(d(32))},
				{"255", t(d(255))},
				{"256", t(d(256))},
				{"16777217", t(d(16777217))},
				{"(+ 1 2)", t(so(), o('+'), d(1), d(2), sc())},
				{"(- 1 2)", t(so(), o('-'), d(1), d(2), sc())},
				{"(* 1 2)", t(so(), o('*'), d(1), d(2), sc())},
				{"(/ 1 2)", t(so(), o('/'), d(1), d(2), sc())},
				{"(/1 2)", t(so(), o('/'), d(1), d(2), sc())},
				{"(/ 1 2 )", t(so(), o('/'), d(1), d(2), sc())},
				{"( / 1 2)", t(so(), o('/'), d(1), d(2), sc())},
				{"(defun doublen (n) (* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc())},
				{"(defun doublen(n) (* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc())},
				{"(defun doublen (n)(* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc())},
				{"(defun doublen(n)(* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc())},
				{"(defun doublen (n) (*n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc())},
				{"(defun doublen (n) (* n 2)   )", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc())},
				{
					"(defun doublen (n) (* n 2))\n (doublen 2)",
					t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o('*'), i("n"), d(2), sc(), sc(), so(), i("doublen"), d(2), sc()),
				},
				{"\"Hello, Coding Challenges\"",
					t(s("\"Hello, Coding Challenges\""))},
				{"(defun hello() (\"Hello Coding Challenges\"))\n(hello)", t(so(), f(), i("hello"), so(), sc(), so(), s("\"Hello Coding Challenges\""), sc(), sc(), so(), i("hello"), sc())},
			},
		),
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

		// evaluate_test("string evaluation", []TestCase[string]{
		// {"\"Hello, Coding Challenges\"",
		// "\"Hello, Coding Challenges\""},
		// {"(defun hello() (\"Hello Coding Challenges\"))\n(hello)",
		// "\"Hello Coding Challenges\""},
		// }),
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
		// ":CC",
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

func so() *Token {
	return &Token{ScopeOpen, []byte{'('}, Unknown}
}

func sc() *Token {
	return &Token{ScopeClose, []byte{')'}, Unknown}
}

func o(ch byte) *Token {
	return &Token{Operator, []byte{ch}, Unknown}
}

func t(args ...*Token) []*Token {
	return args
}

func d(i int) *Token {
	return &Token{Literal, int32_to_bytes(i), Int}
}

func f() *Token {
	return &Token{Keyword, []byte("defun"), Unknown}
}

func i(s string) *Token {
	return &Token{Identifier, []byte(s), Unknown}
}

func s(s string) *Token {
	return &Token{Literal, []byte(s), String}
}

func eq(a []*Token, b []*Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if e.TokenKind != b[i].TokenKind || e.Type != b[i].Type {
			return false
		}
		for j, ch := range e.Value {
			if ch != b[i].Value[j] {
				return false
			}
		}
	}

	return true
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
