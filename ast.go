package main

import (
	"fmt"
	"slices"
)

type Parser struct {
	Functions *[]FunctionEntry
	Errors    *[]error
	Position  int
	Input     string
}

type Ast struct {
	Functions *[]FunctionEntry
	Root      *Node
}

func ParseAst(input string) (Ast, []error) {
	p := Parser{&[]FunctionEntry{}, &[]error{}, 0, input}
	node := p.atom_node()

	return Ast{p.Functions, node}, *p.Errors
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

func coalesce(nodes ...func() *Node) *Node {
	for _, f := range nodes {
		n := f()
		if n != nil {
			return n
		}
	}
	return nil
}

func (p *Parser) atom_node() *Node {
	p.trim()
	return coalesce(
		// atom can be an expression consisting of other atoms/expressions, it can contain strings and other stuff from below
		func() *Node { return p.expression_node() },
		// atom can be a string, string can contain other stuff from below
		func() *Node { return p.string_node() },
		// can handle more corner cases here
		// eat value and extract atom
		func() *Node { return p.eat_atom() },
	)
}

func (p *Parser) eat_atom() *Node {
	val := p.eat_value()
	return coalesce(
		func() *Node { return int_node(val) },
		func() *Node { return bool_node(val) },
		func() *Node { return decl_node(val) },
		func() *Node { return p.binary_operator_node(val) },
		// fallback
		func() *Node { return node(Atom, []byte(val), Unknown, []*Node{}) },
	)
}

func (p *Parser) binary_operator_node(s string) *Node {
	if len(s) == 1 && s[0] == '+' || s[0] == '-' || s[0] == '*' || s[0] == '/' {
		return node(BinaryOperator, []byte(s), Unknown, []*Node{})
	}
	return nil
}

func (p *Parser) expression_node() *Node {
	p.trim()
	if p.is('(') {
		p.drop()
		nodes := []*Node{}
		types := []ResultType{}
		for p.isnt(')') {
			// expression consists of atoms which can be other expressions, strings or basic atoms
			node := p.atom_node()
			if node != nil {
				nodes = append(nodes, node)
				if node.Type != Unknown {
					types = append(types, node.Type)
				}
				// function declaration case
				if node.Kind == DeclarationAtom {
					// declaration := node
					// extract name
					node = p.atom_node()
					node.Kind = SymbolAtom
					nodes = append(nodes, node)
					arguments, body := p.extract_function()
					if arguments != nil && body != nil {
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
			return node(Expression, []byte{}, types[0], nodes)
		}
		return node(Expression, []byte{}, Unknown, nodes)
	}
	return nil
}

func (p *Parser) extract_function() (*Node, *Node) {
	// extract arguments
	arguments := p.expression_node()

	// fmt.Printf("\nextracted arguments: %v", arguments)
	if arguments != nil {
		arguments.Kind = ArgumentsExpression
		args := []string{}
		for _, arg := range arguments.Nodes {
			arg.Kind = ArgumentDeclaration
			args = append(args, string(arg.Value))
		}

		// extract body
		body := p.expression_node()
		fmt.Printf("\nextracted body: %v", body)
		if body == nil {
			return nil, nil
		}

		// reference beteween the nodes inside of the body to the arguments
		p.setup_argument_references(body, arguments)
		return arguments, body
	}
	return nil, nil
}

func (p Parser) record_error(err error) {
	*p.Errors = append(*p.Errors, err)
}

func (p Parser) setup_argument_references(node *Node, arguments *Node) {
	for _, n := range node.Nodes {
		for i, a := range arguments.Nodes {
			// the argument name corresponds the atom
			if slices.Equal(a.Value, n.Value) {
				n.Kind = ArgumentVariable
				// fmt.Printf("\nargument reference setup for %v to %v", a, n)
				n.ArgumentIndex = i
				if a.Type == Unknown {
					a.Type = n.Type
				} else {
					if a.Type != n.Type {
						p.record_error(fmt.Errorf("Inconsistent typing of argument %v", string(a.Value)))
					}
				}
			}
		}
		if len(n.Nodes) > 0 {
			p.setup_argument_references(n, arguments)
		}
	}
}

func (p *Parser) string_node() *Node {
	if p.is('"') {
		begin := p.Position
		// include open quote
		p.drop()
		for p.isnt('"') {
			p.drop()
		}
		// include closing quote
		p.drop()
		return node(Atom, []byte(p.Input[begin:p.Position]), String, []*Node{})
	}
	return nil
}
