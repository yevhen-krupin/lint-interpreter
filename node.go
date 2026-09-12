package main

import "log"

// "fmt"

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
	Expression       Kind = "Expression"
	Atom             Kind = "Atom"
	ConditionAtom    Kind = "ConditionAtom"
	SymbolAtom       Kind = "SymbolAtom"
	BinaryOperator   Kind = "BinaryOperator"
	ArgumentVariable Kind = "ArgumentVariable"
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

	log.Println("node", &node, ": kind ", kind, " | type ", rt, "| value ", string(value), " | nodes", nodes)
	return &node
}
