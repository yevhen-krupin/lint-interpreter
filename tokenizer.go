package main

import (
	"fmt"
	"iter"
	"slices"
	"strings"
	"unicode"
)

type TokenKind string

const (
	Identifier  TokenKind = "Identifier"
	Operator    TokenKind = "Operator"
	ScopeOpen   TokenKind = "ScopeOpen"
	ScopeClose  TokenKind = "ScopeClose"
	Literal     TokenKind = "Literal"
	Keyword     TokenKind = "Keyword"
	Declaration TokenKind = "Declaration"
)

type TokenStream []*Token

func (s TokenStream) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, e := range s {
		sb.WriteString(string(e.TokenKind))
		sb.WriteRune(' ')
		sb.Write(e.Value)
		sb.WriteRune(' ')
		sb.WriteString(string(e.Type))
		if i < len(s)-1 {
			sb.WriteRune(',')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

type Token struct {
	TokenKind TokenKind
	Value     []byte
	Type      ResultType
}

var operators = []byte("+-/*")

type Tokenizer struct {
	Input    string
	Position int
}

func tokenize(input string) iter.Seq[*Token] {
	tokenizer := &Tokenizer{input, 0}
	return tokenizer.tokenize0()
}

func (tokenizer *Tokenizer) tokenize0() iter.Seq[*Token] {
	return func(yield func(*Token) bool) {
		for tokenizer.in() {
			tokenizer.trim()
			fmt.Printf("\ninput left: %v", tokenizer.Input[tokenizer.Position:len(tokenizer.Input)])
			t := coalesce(
				tokenizer,
				(*Tokenizer).scope,
				(*Tokenizer).operator,
				(*Tokenizer).declaration,
				(*Tokenizer).literal,
				(*Tokenizer).identifier,
			)
			if t != nil {
				if !yield(t) {
					return
				}
			}
		}
	}
}

func (p *Tokenizer) scope() *Token {
	if p.is('(') {
		return &Token{ScopeOpen, p.drop(1), Unknown}
	}
	if p.is(')') {
		return &Token{ScopeClose, p.drop(1), Unknown}
	}
	return nil
}

func (p *Tokenizer) operator() *Token {
	if p.is('+') || p.is('-') || p.is('*') || p.is('/') {
		return &Token{Operator, p.drop(1), Unknown}
	}
	return nil
}

func (p *Tokenizer) declaration() *Token {
	// when check include whitespace
	if p.are("defun ") {
		return &Token{Declaration, p.drop(len("defun")), Unknown}
	}
	return nil
}

func (p *Tokenizer) literal() *Token {
	if p.is('"') {
		length := p.whilent([]byte{'"'})
		return &Token{Literal, p.drop(length + 1), String}
	}
	if p.are("t ") || p.are("T ") {
		return &Token{Literal, p.drop(1), Boolean}
	}
	if p.are("nil ") {
		return &Token{Literal, p.drop(len("nil")), Boolean}
	}
	// are numbers = literal
	digits := p.while(unicode.IsDigit)
	if digits > 0 {
		// fmt.Printf("\n[token] %d %v", digits, string(p.Input[p.Position:p.Position+digits]))
		return &Token{Literal, p.drop(digits), Int}
	}
	return nil
}

func (p *Tokenizer) identifier() *Token {
	// starts from letter, can contain letters or numbers or underscores
	letters := p.while(unicode.IsLetter)
	if letters > 0 {
		length := p.while(func(r rune) bool {
			return unicode.IsDigit(r) || unicode.IsLetter(r)
		})
		return &Token{Identifier, p.drop(length), Unknown}
	}
	return nil
}

func (p *Tokenizer) in() bool {
	return p.Position < len(p.Input)
}

func (p *Tokenizer) while(f func(rune) bool) int {
	i := p.Position
	for i < len(p.Input) && f(rune(p.Input[i])) == true {
		i += 1
	}
	return i - p.Position
}

func (p *Tokenizer) are(chars string) bool {
	for i := 0; i < len(chars) && p.Position+i < len(p.Input); i++ {
		if p.Input[p.Position+i] != chars[i] {
			return false
		}
	}
	return true
}

func (p *Tokenizer) whilent(chars []byte) int {
	i := p.Position
	for i < len(p.Input) && slices.Index(chars, p.Input[i]) == -1 {
		i += 1
	}
	return i - p.Position
}

func (p *Tokenizer) isany(chars []byte) bool {
	return p.in() && slices.Contains(chars, p.Input[p.Position])
}

func (p *Tokenizer) is(char byte) bool {
	return p.in() && p.Input[p.Position] == char
}

func (p *Tokenizer) isnt(char byte) bool {
	return p.in() && !p.is(char)
}

func (p *Tokenizer) drop(count int) []byte {
	s := []byte(p.Input[p.Position : p.Position+count])
	p.Position += count
	return s
}

func (p *Tokenizer) trim() {
	// tokens can be split by spaces or even new lines
	for p.is(' ') || p.is('\n') {
		p.drop(1)
	}
}
