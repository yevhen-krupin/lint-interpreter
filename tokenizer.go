package main

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type TokenKind string

const (
	Identifier TokenKind = "Identifier"
	Operator   TokenKind = "Operator"
	ScopeOpen  TokenKind = "ScopeOpen"
	ScopeClose TokenKind = "ScopeClose"
	Literal    TokenKind = "Literal"
	Keyword    TokenKind = "Keyword"
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
		last_pos := tokenizer.Position
		for tokenizer.in() {
			tokenizer.trim()
			//fmt.Printf("\ninput left: %v", tokenizer.Input[tokenizer.Position:len(tokenizer.Input)])
			t := coalesce(
				tokenizer,
				(*Tokenizer).scope,
				(*Tokenizer).operator,
				(*Tokenizer).declaration,
				(*Tokenizer).condition,
				(*Tokenizer).literal,
				(*Tokenizer).identifier,
			)
			if t != nil {
				if !yield(t) {
					return
				}
			} else {
				if tokenizer.Position == last_pos {
					panic(fmt.Sprintf("\ntokenizer is stuck, unknown token %v", tokenizer.Input[tokenizer.Position:len(tokenizer.Input)]))
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
		return &Token{Operator, p.drop(1), Int}
	}
	if p.is('<') || p.is('>') || p.is('=') || p.are("<=") || p.are(">=") {
		return &Token{Operator, p.drop(1), Boolean}
	}
	return nil
}

func (p *Tokenizer) condition() *Token {
	if p.are("if ") || p.are("if(") {
		return &Token{Keyword, p.drop(2), Unknown}
	}
	return nil
}

func (p *Tokenizer) declaration() *Token {
	// when check include whitespace
	if p.are("defun ") {
		return &Token{Keyword, p.drop(len("defun")), Unknown}
	}
	return nil
}

func (p *Tokenizer) literal() *Token {
	if p.is('"') {
		p.drop(1)
		length := p.whilent([]byte{'"'})
		return &Token{Literal, slices.Concat([]byte{'"'}, p.drop(length+1)), String}
	}
	if p.are("t ") || p.are("T ") {
		p.drop(1)
		return &Token{Literal, []byte{1}, Boolean}
	}
	if p.are("nil ") {
		p.drop(3)
		return &Token{Literal, []byte{0}, Boolean}
	}
	// are numbers = literal
	digits := p.while(unicode.IsDigit)
	if digits > 0 {
		// convert bytes -> string -> number
		s := string(p.drop(digits))
		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Printf("\nError: unable to convert %v to int %v", s, err)
		}
		// fmt.Printf("\n[token] %d %v", digits, string(p.Input[p.Position:p.Position+digits]))
		return &Token{Literal, int32_to_bytes(i), Int}
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
