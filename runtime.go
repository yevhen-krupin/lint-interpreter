package main

import (
	"fmt"

	"github.com/google/uuid"
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
	// condition
	if node.Kind == ConditionAtom {
		if len(node.Nodes) != 3 {
			return EvaluationResult{Unknown, nil, fmt.Errorf("condition atom should have 3 children nodes: condition body and two branches for true and false")}
		}
		if node.Type != Boolean {
			return EvaluationResult{Unknown, nil, fmt.Errorf("condition atom have boolean type")}
		}
		result := r.evaluate(node.Nodes[0])
		fmt.Println("evaluate condition", value_to_string(node.Nodes[0].Nodes[0]), value_to_string(node.Nodes[0].Nodes[1]), value_to_string(node.Nodes[0].Nodes[2]), result.Type, result.Value)
		if result.Error != nil {
			return result
		}
		if result.Type != Boolean {
			return EvaluationResult{Unknown, nil, fmt.Errorf("condition evaluation should to return boolean but was %v : %v", result.Type, result.Value)}
		}
		if result.Value == true {
			fmt.Println("condition -> true branch")
			print(node.Nodes[1], "")
			return r.evaluate(node.Nodes[1])
		} else {
			fmt.Println("condition -> false branch", node.Nodes[2])
			print(node.Nodes[2], "")
			return r.evaluate(node.Nodes[2])
		}
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
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) + b.Value.(int), nil}
			})
		}
		if oper[0] == '-' {
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) - b.Value.(int), nil}
			})
		}
		if oper[0] == '*' {
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) * b.Value.(int), nil}
			})
		}
		if oper[0] == '/' {
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) / b.Value.(int), nil}
			})
		}
		if oper[0] == '=' {
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Boolean, a.Value.(int) == b.Value.(int), nil}
			})
		}
		if oper[0] == '<' {
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				//fmt.Println("evaluate <", a.Value.(int), b.Value.(int), a.Value.(int) < b.Value.(int))
				return EvaluationResult{Boolean, a.Value.(int) < b.Value.(int), nil}
			})
		}
		if oper[0] == '>' {
			return r.binary_operator(node, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Boolean, a.Value.(int) > b.Value.(int), nil}
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
	if len(node.Nodes) < 1 && node.Nodes[0].Kind == BinaryOperator {
		return []byte{}, fmt.Errorf("unable to get operator atom %v", node.Value)
	}
	return node.Nodes[0].Value, nil
}

func (r Runtime) binary_operator(node *Node, f func(EvaluationResult, EvaluationResult) EvaluationResult) EvaluationResult {
	a := r.operand(node, 1)
	b := r.operand(node, 2)
	if a.Error != nil || b.Error != nil {
		return EvaluationResult{Error, nil, first_error(a.Error, b.Error)}
	}
	return f(a, b)
}

func (r Runtime) operand(node *Node, index int) EvaluationResult {
	return r.evaluate(node.Nodes[index])
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

	fmt.Println("calling function", name, "args", args, "body", function.Body)
	frame := StackFrame{uuid.New().String(), name, args}
	*r.Stack = append(*r.Stack, frame)
	res := r.evaluate(function.Body)
	top := (*r.Stack)[len(*r.Stack)-1]
	if top.id != frame.id {
		panic("unexpected state: the stack frame after leaving the function is not correct")
	}
	*r.Stack = (*r.Stack)[:len(*r.Stack)-1]
	fmt.Println("function", name, "args", args, "result", res.Type, ":", res.Value)
	return res
}
