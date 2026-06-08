package main

import (
	"fmt"
)

type Runtime struct {
	Stack     *[]StackFrame
	Functions *[]FunctionEntry
}

func NewRuntime() Runtime {
	return Runtime{&[]StackFrame{}, &[]FunctionEntry{}}
}

func (r Runtime) EvaluateAst(ast Ast) EvaluationResult {
	*r.Functions = append(*r.Functions, *ast.Functions...)
	*r.Stack = []StackFrame{{Function: "__entry_point", Arguments: []Argument{}}}
	return r.evaluate(ast.Root)
}

func (r Runtime) last_frame() StackFrame {
	return (*r.Stack)[len(*r.Stack)-1]
}

func (r Runtime) evaluate(node *Node) EvaluationResult {
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

	// resolve argument variable
	if node.Kind == ArgumentVariable && len(node.Nodes) == 0 && node.ArgumentIndex >= 0 && len(*r.Stack) > 0 && node.ArgumentIndex < len(r.last_frame().Arguments) {
		// todo: convert into a real stack
		arg_value := r.last_frame().Arguments[node.ArgumentIndex]
		return EvaluationResult{arg_value.Type, arg_value.Value, nil}
	}
	// call function
	if node.Kind == Expression && len(node.Nodes) > 0 && node.Nodes[0].Kind == Atom {
		for _, f := range *r.Functions {
			if f.Name == string(node.Nodes[0].Value) {
				return r.call_function(node, f)
			}
		}
	}

	// operators
	if node.Kind == Expression && len(node.Nodes) > 2 {
		oper, err := r.get_operator_bytes(node)
		if err != nil {
			return EvaluationResult{Error, nil, err}
		}
		if oper[0] == '+' {
			return r.binary_operator_int(node, func(a, b int) int {
				return a + b
			})
		}
		if oper[0] == '-' {
			return r.binary_operator_int(node, func(a, b int) int {
				return a - b
			})
		}
		if oper[0] == '*' {
			return r.binary_operator_int(node, func(a, b int) int {
				return a * b
			})
		}
		if oper[0] == '/' {
			return r.binary_operator_int(node, func(a, b int) int {
				return a / b
			})
		}
	}
	// expression that just return something
	if node.Kind == Expression && len(node.Nodes) == 1 {
		return r.evaluate(node.Nodes[0])
	}

	return EvaluationResult{Unknown, nil, nil}
}

func (r Runtime) get_operator_bytes(node *Node) ([]byte, error) {
	if len(node.Nodes) < 1 /*|| node.Nodes[0].Kind != Atom*/ {
		return []byte{}, fmt.Errorf("unable to get operator atom %v", node.Value)
	}
	return node.Nodes[0].Value, nil
}

func (r Runtime) binary_operator_int(node *Node, f func(int, int) int) EvaluationResult {
	a, e1 := r.operand_int(node, 1)
	b, e2 := r.operand_int(node, 2)
	if e1 != nil || e2 != nil {
		return EvaluationResult{Error, nil, first_error(e1, e2)}
	}
	return EvaluationResult{Int, f(a, b), nil}
}

func (r Runtime) operand_int(node *Node, index int) (int, error) {
	res := r.evaluate(node.Nodes[index])
	if res.Type == Int {
		return res.Value.(int), res.Error
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

// node: symbol atom of call site
// first child is a function, followed by the arguments
func (r Runtime) call_function(node *Node, function FunctionEntry) EvaluationResult {
	name := string(node.Nodes[0].Value)
	args := []Argument{}
	for i := 1; i < len(node.Nodes); i++ {
		res := r.evaluate(node.Nodes[i])
		if res.Error != nil {
			//todo: probably worth wrapping with more context
			return res
		}
		arg := Argument{res.Value, res.Type}
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
	*r.Stack = append(*r.Stack, StackFrame{name, args})
	return r.evaluate(function.Body)
}
