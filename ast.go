package main

import (
	"fmt"
	"slices"
)

type Parser struct {
	Functions *[]FunctionEntry
	Errors    []error
	Position  int
	Input     string
}

type Ast struct {
	Functions *[]FunctionEntry
	Root      *Node
}

func ParseAst(input string) (Ast, []error) {
	p := Parser{&[]FunctionEntry{}, []error{}, 0, input}
	_, node := p.atom_node()

	return Ast{p.Functions, node}, []error{}
}

func (p *Parser) in() bool {
	return p.Position < len(p.Input)
}

func (p *Parser) is(char byte) bool {
	return p.in() && p.Input[p.Position] == char
}

func (p *Parser) isnt(char byte) bool {
	return p.in() && !p.is(char)
}

func (p *Parser) drop() {
	p.Position += 1
}

func (p *Parser) trim() {
	// atoms are split by spaces
	for p.is(' ') {
		p.drop()
	}
}

func (p *Parser) eat_value() string {
	begin := p.Position
	// get all while it is not a separator or an end of current expression or a begin of a new one
	for p.isnt(' ') && p.isnt(')') && p.isnt('(') {
		p.drop()
	}

	return p.Input[begin:p.Position]
}

type NodeFunction func() *Node

// todo
// func coalesce(funcs ...func() *Node) *Node {
// for f := range funcs {
// n := f()
// }
// }

func (p *Parser) atom_node() (bool, *Node) {
	p.trim()
	// atom can be an expression consisting of other atoms/expressions
	r, n := p.expression_node()
	if r {
		return r, n
	}
	// atom can be a string atom
	r, n = p.string_node()
	if r {
		return r, n
	}
	// can handle more corner cases here
	// extract atom
	val := p.eat_value()
	r, n = int_node(val)
	if r {
		return true, n
	}
	r, n = bool_node(val)
	if r {
		return true, n
	}
	r, n = decl_node(val)
	if r {
		return true, n
	}
	return true, node(Atom, []byte(val), Unknown, []*Node{})
}

func (p *Parser) expression_node() (bool, *Node) {
	p.trim()
	if p.is('(') {
		p.drop()
		nodes := []*Node{}
		types := []ResultType{}
		for p.isnt(')') {
			// expression consists of atoms which can be other expressions, strings or basic atoms
			r, node := p.atom_node()
			if r {
				nodes = append(nodes, node)
				if node.Type != Unknown {
					types = append(types, node.Type)
				}
				// function declaration case
				if node.Kind == DeclarationAtom {
					// declaration := node
					// extract name
					r, node = p.atom_node()
					node.Kind = SymbolAtom
					nodes = append(nodes, node)
					r2, arguments, body := p.extract_function()
					if r2 {
						// name symbol return type (function return type) corresponds body
						node.Type = body.Type
						nodes = append(nodes, arguments)
						nodes = append(nodes, body)
						*p.Functions = append(*p.Functions, FunctionEntry{Name: string(node.Value), Body: body, Arguments: arguments})

					}
				}
			}
		}
		p.drop()
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
			return true, node(Expression, []byte{}, types[0], nodes)
		}
		return true, node(Expression, []byte{}, Unknown, nodes)
	}
	return false, nil
}

func (p *Parser) extract_function() (bool, *Node, *Node) {
	// extract arguments
	r, arguments := p.expression_node()

	fmt.Printf("\nextracted arguments: %v", arguments)
	if r {
		arguments.Kind = ArgumentsExpression
		args := []string{}
		for _, arg := range arguments.Nodes {
			arg.Kind = SymbolAtom
			args = append(args, string(arg.Value))
		}

		// extract body
		r2, body := p.expression_node()
		fmt.Printf("\nextracted body: %v", body)
		if r2 == false {
			return false, nil, nil
		}

		// reference beteween the nodes inside of the body to the arguments
		// todo: more depth
		for _, n := range body.Nodes {
			reference := slices.Index(args, string(n.Value))
			if reference > -1 {
				n.Kind = SymbolAtom
				n.ArgumentIndex = reference
				arguments.Nodes[reference].Type = n.Type
			}
		}
		return true, arguments, body
	}
	return false, nil, nil
}

func (p *Parser) string_node() (bool, *Node) {
	if p.is('"') {
		begin := p.Position
		// include open quote
		p.drop()
		for p.isnt('"') {
			p.drop()
		}
		// include closing quote
		p.drop()
		return true, node(Atom, []byte(p.Input[begin:p.Position]), String, []*Node{})
	}
	return false, nil
}
