package main

import (
	"fmt"
	"slices"
)

type Parser struct {
	Functions *[]FunctionEntry
	Errors    *[]error
	Position  int
	Input     []*Token
}

type Ast struct {
	Functions *[]FunctionEntry
	Root      *Node
}

func ParseAst(input string) (Ast, []error) {
	p := Parser{&[]FunctionEntry{}, &[]error{}, 0, slices.Collect(tokenize(input))}
	node := p.atom_node()

	return Ast{p.Functions, node}, *p.Errors
}

func (p *Parser) atom_node() *Node {
	return coalesce(
		p,
		// atom can be an expression consisting of other atoms/expressions, it can contain strings and other stuff from below
		(*Parser).expression_node,
		// atom can be a literal: string, int, boolean
		(*Parser).literal_node,

		// extract declaration
		(*Parser).decl_node,

		// extract binary operators
		(*Parser).binary_operator_node,
		//fallback
		(*Parser).undefined_node,
	)
}

func (p *Parser) expression_node() *Node {
	if p.peek().TokenKind == ScopeOpen {
		p.eat()
		nodes := []*Node{}
		types := []ResultType{}

		for p.peek().TokenKind != ScopeClose {
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
		p.eat()
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

func (p *Parser) literal_node() *Node {
	if p.peek().TokenKind == Literal {
		return node(Atom, p.peek().Value, p.eat().Type, []*Node{})
	}
	return nil
}

func (p *Parser) decl_node() *Node {
	if p.peek().TokenKind == Keyword && p.eq(p.peek().Value, "defun") {
		return node(DeclarationAtom, p.eat().Value, Unknown, []*Node{})
	}
	return nil
}

func (p *Parser) binary_operator_node() *Node {
	if p.peek().TokenKind == Operator {
		return node(BinaryOperator, p.eat().Value, Unknown, []*Node{})
	}
	return nil
}

func (p *Parser) peek() *Token {
	if !p.in() {
		return &Token{}
	}
	return p.Input[p.Position]
}

func (p *Parser) undefined_node() *Node {
	return node(Atom, p.eat().Value, Unknown, []*Node{})
}

func (p Parser) record_error(err error) {
	*p.Errors = append(*p.Errors, err)
}

func (p *Parser) in() bool {
	return p.Position < len(p.Input)
}

func (p *Parser) eat() *Token {
	p.Position += 1
	return p.Input[p.Position-1]
}

func (p *Parser) eq(bytes []byte, str string) bool {
	if len(bytes) != len(str) {
		return false
	}
	for i, ch := range bytes {
		if str[i] != ch {
			return false
		}
	}
	return true
}
