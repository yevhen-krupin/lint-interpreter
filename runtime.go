package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

type Runtime struct {
	Stack       *[]StackFrame
	SymbolTable *SymbolTable
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
	id        string
	Function  string
	Arguments []*EvaluationResult
}

type FunctionEntry struct {
	Name      string
	Body      *Node
	Arguments *Node
}

func NewRuntime(symbolTable *SymbolTable) Runtime {
	return Runtime{&[]StackFrame{}, symbolTable}
}

func (r Runtime) EvaluateAst(ast *Node) EvaluationResult {
	*r.Stack = []StackFrame{{Function: "__entry_point", Arguments: []*EvaluationResult{}}}
	return r.evaluate(ast)
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
		result := r.evaluate(node.Nodes[0])
		if result.Error != nil {
			return result
		}
		if result.Type != Boolean {
			return EvaluationResult{Unknown, nil, fmt.Errorf("condition evaluation should to return boolean but was %v : %v", result.Type, result.Value)}
		}
		if result.Value == true {
			log.Println("condition -> true branch")
			return r.evaluate(node.Nodes[1])
		} else {
			log.Println("condition -> false branch", node.Nodes[2])
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
	if node.Kind == Expression && len(node.Nodes) > 0 && node.Nodes[0].Kind == SymbolAtom {
		if f, ok := (*r.SymbolTable.Functions)[string(node.Nodes[0].Value)]; ok {
			resolved := []*EvaluationResult{}
			for i, n := range node.Nodes {
				if i > 0 {
					arg := r.evaluate(n)
					log.Println("resolved argument", n, arg)
					resolved = append(resolved, &arg)
				}
			}
			return r.call_function(string(node.Nodes[0].Value), resolved, f.Arguments, f.Body)
		}
	}

	// operators
	if node.Kind == BinaryOperator {
		oper := node
		if oper.Value[0] == '+' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) + b.Value.(int), nil}
			})
		}
		if oper.Value[0] == '-' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) - b.Value.(int), nil}
			})
		}
		if oper.Value[0] == '*' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) * b.Value.(int), nil}
			})
		}
		if oper.Value[0] == '/' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Int, a.Value.(int) / b.Value.(int), nil}
			})
		}
		if oper.Value[0] == '=' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Boolean, a.Value.(int) == b.Value.(int), nil}
			})
		}
		if oper.Value[0] == '<' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				//log.Println("evaluate <", a.Value.(int), b.Value.(int), a.Value.(int) < b.Value.(int))
				return EvaluationResult{Boolean, a.Value.(int) < b.Value.(int), nil}
			})
		}
		if oper.Value[0] == '>' {
			return r.binary_operator(oper, func(a, b EvaluationResult) EvaluationResult {
				return EvaluationResult{Boolean, a.Value.(int) > b.Value.(int), nil}
			})
		}
	}
	// expression that just returns something
	if node.Kind == Expression && len(node.Nodes) == 1 {
		return r.evaluate(node.Nodes[0])
	}

	return EvaluationResult{Unknown, nil, fmt.Errorf("unrecognized node %v", node)}
}

func (r Runtime) binary_operator(node *Node, f func(EvaluationResult, EvaluationResult) EvaluationResult) EvaluationResult {
	log.Println("evaluating binary operator", node)
	a := r.operand(node, 0)
	b := r.operand(node, 1)
	if a.Error != nil || b.Error != nil {
		return EvaluationResult{Error, nil, first_error(a.Error, b.Error)}
	}
	return f(a, b)
}

func (r Runtime) operand(node *Node, index int) EvaluationResult {
	log.Println("evaluating operand", node.Nodes[index])
	return r.evaluate(node.Nodes[index])
}

// node: symbol atom of call site
// first child is a function, followed by the arguments
func (r Runtime) call_function(name string, args []*EvaluationResult, functionArguments *Node, functionBody *Node) EvaluationResult {
	// check arguments
	if len(functionArguments.Nodes) != len(args) {
		return EvaluationResult{Error: fmt.Errorf("arguments count provided to the function `%v` don't match declared function, expected: %d, actual: %d", name, len(functionArguments.Nodes), len(args))}
	}
	for i := 0; i < len(args); i++ {
		if args[i].Type != functionArguments.Nodes[i].Type {
			return EvaluationResult{Error: fmt.Errorf("the argument %v type provided to the function `%v` doesn't match declared function's argument type, expected: %v, actual: %v", string(functionArguments.Nodes[i].Value), name, functionArguments.Nodes[i].Type, args[i].Type)}
		}
	}

	log.Println("calling function", name, "args", args, "body", functionBody)
	frame := StackFrame{uuid.New().String(), name, args}
	*r.Stack = append(*r.Stack, frame)
	res := r.evaluate(functionBody)
	top := (*r.Stack)[len(*r.Stack)-1]
	if top.id != frame.id {
		panic("unexpected state: the stack frame after leaving the function is not correct")
	}
	*r.Stack = (*r.Stack)[:len(*r.Stack)-1]
	log.Println("function", name, "args", args, "result", res.Type, ":", res.Value)
	return res
}

func first_error(input ...error) error {
	for _, item := range input {
		if item != nil {
			return item
		}
	}
	return nil
}
