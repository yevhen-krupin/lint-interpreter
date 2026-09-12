package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const TEST_DIR = "build/test/work"

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

func evaluate_test[T int | bool | string](name string, array []TestCase[T]) TestResult {
	pass := 0
	fail := 0
	for index, element := range array {
		if evaluate_test_case(name, index, element) {
			passed(element.input)
			pass += 1
		} else {
			failed(element.input)
			fail += 1
		}
	}
	report(name, pass, fail)
	return TestResult{pass, fail}
}

func report(name string, pass int, fail int) {
	log.Println("[", name, "] result:", Green, "Pass", pass, "|", Red, "Fail", fail, "|", Reset, "Total", pass+fail)
}

func passed(input string) {
	log.Println(Green, "PASS", Reset, "`", strings.ReplaceAll(strings.ReplaceAll(input, "  ", " "), "\n", " "), "`")
}

func failed(input string) {
	log.Println(Red, "FAIL", Reset, "`", input, "`")
}

type Closeable interface {
	Close()
}

type TestScope struct {
	f *os.File
}

func (t TestScope) Close() {
	t.f.Close()
	log.SetOutput(os.Stdout)
}

func wrap_test(index int, name string, input string) Closeable {
	if _, err := os.Stat(TEST_DIR); err != nil {
		if os.IsNotExist(err) {
			if os.MkdirAll(TEST_DIR, 0777) != nil {
				log.Fatal("error making a dir")
			}
		}
	}
	f, err := os.OpenFile(fmt.Sprintf("%v/%v_%v.txt", TEST_DIR, name, index), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	return TestScope{f}
}

func evaluate_test_case[T int | bool | string](name string, index int, element TestCase[T]) bool {
	results := []EvaluationResult{}
	nodes := []*Node{}
	subs := strings.SplitSeq(element.input, "\n")
	runtime := NewRuntime()
	s := wrap_test(index, name, element.input)
	defer s.Close()
	for sub_element := range subs {
		ast, _ := ParseAst(sub_element)
		nodes = append(nodes, ast.Root)
		r := runtime.EvaluateAst(ast)
		results = append(results, r)
		if r.Error != nil {
			log.Println("Failed to evaluate due to", r.Error.Error())
		}
	}
	final := results[len(results)-1]

	log.Println("Functions declared:", runtime.Functions)
	for i, r := range results {
		if i == len(results)-1 {
			log.Println("Evaluated as", r.Type, r.Value, " but has to be", element.output, element.output)
		} else {
			log.Println("Evaluated as:", r.Type, r.Value)
		}
		print(nodes[i], "")
	}

	return final.Value == element.output
}
