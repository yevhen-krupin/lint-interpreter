package main

import (
// "fmt"
)

type ResultType string

const (
	Boolean ResultType = "Boolean"
	String  ResultType = "String"
	Int     ResultType = "Int"
	Error   ResultType = "Error"
	Unknown ResultType = "Unknown"
)

type Kind string

// todo: add more kinds: argument, body, call site
const (
	Expression          Kind = "Expression"
	Atom                Kind = "Atom"
	DeclarationAtom     Kind = "DeclarationAtom"
	SymbolAtom          Kind = "SymbolAtom"
	ArgumentsExpression Kind = "ArgumentsExpression"
	BinaryOperator      Kind = "BinaryOperator"
	ArgumentDeclaration Kind = "ArgumentDeclaration"
	ArgumentVariable    Kind = "ArgumentVariable"
)

type Node struct {
	Kind          Kind
	Type          ResultType
	Value         []byte
	Nodes         []*Node
	ArgumentIndex int
}

func node(kind Kind, value []byte, rt ResultType, nodes []*Node) *Node {
	var node Node
	node.Type = rt
	node.Kind = kind
	node.Value = value
	node.Nodes = nodes

	//fmt.Printf("\nnode %p: kind %v | type %v | value %v | nodes %v", &node, kind, rt, string(value), nodes)
	return &node
}
