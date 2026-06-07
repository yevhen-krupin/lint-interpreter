package main

import (
	//"fmt"
	"strconv"
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

const (
	Expression          Kind = "Expression"
	Atom                Kind = "Atom"
	DeclarationAtom     Kind = "DeclarationAtom"
	SymbolAtom          Kind = "SymbolAtom"
	ArgumentsExpression Kind = "ArgumentsExpression"
)

type Node struct {
	Kind          Kind
	Type          ResultType
	Value         []byte
	Nodes         []*Node
	ArgumentIndex int
}

func int_node(input string) (bool, *Node) {
	i, err := strconv.Atoi(input)
	if err == nil {
		return true, node(Atom, *int32_to_bytes(i), Int, []*Node{})
	}
	return false, nil
}

func bool_node(input string) (bool, *Node) {
	if input == "t" || input == "T" {
		return true, node(Atom, []byte{1}, Boolean, []*Node{})
	}
	if input == "nil" {
		return true, node(Atom, []byte{0}, Boolean, []*Node{})
	}
	return false, nil
}

func decl_node(input string) (bool, *Node) {
	if input == "defun" {
		return true, node(DeclarationAtom, []byte(input), Unknown, []*Node{})
	}
	return false, nil
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
