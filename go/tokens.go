package main

import (
	"fmt"
	"strings"
)

var keywords = map[string]bool{
	"function": true,
	"return":   true,
	"object":   true,
	"new":      true,
	"free":     true,
}

type token struct {
	is    string
	value string
}

type tokenizer struct {
	stream  *stream
	current string
}

func (me *token) string() string {
	if me.value == "" {
		return fmt.Sprintf("{type:%s}", me.is)
	}
	return fmt.Sprintf("{type:%s, value:%s}", me.is, me.value)
}

func simpleToken(is string) *token {
	t := &token{}
	t.is = is
	t.value = ""
	return t
}

func valueToken(is, value string) *token {
	t := &token{}
	t.is = is
	t.value = value
	return t
}

func digit(c byte) bool {
	return strings.IndexByte("0123456789", c) >= 0
}

func letter(c byte) bool {
	return strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c) >= 0
}
func (me *tokenizer) forSpace() {
	stream := me.stream
	for !stream.eof() {
		c := stream.peek()
		if c != ' ' || c == '\t' {
			break
		}
		stream.next()
	}
}

func (me *tokenizer) forNumber() string {
	stream := me.stream
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.peek()
		if !digit(c) {
			break
		}
		value.WriteByte(c)
		stream.next()
	}
	return value.String()
}

func (me *tokenizer) forWord() string {
	stream := me.stream
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.peek()
		if !letter(c) {
			break
		}
		value.WriteByte(c)
		stream.next()
	}
	return value.String()
}

func (me *tokenizer) forString() string {
	stream := me.stream
	stream.next()
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.next()
		if c == '"' {
			break
		}
		value.WriteByte(c)
	}
	return value.String()
}

func tokenize(stream *stream) []*token {
	me := tokenizer{}
	me.stream = stream
	tokens := make([]*token, 0)
	size := len(stream.data)
	for stream.pos < size {
		me.forSpace()
		number := me.forNumber()
		if number != "" {
			token := valueToken("number", number)
			fmt.Println(len(tokens), token.string())
			tokens = append(tokens, token)
			continue
		}
		word := me.forWord()
		if word != "" {
			var token *token
			if _, ok := keywords[word]; ok {
				token = simpleToken(word)
			} else {
				token = valueToken("id", word)
			}
			fmt.Println(len(tokens), token.string())
			tokens = append(tokens, token)
			continue
		}
		c := stream.peek()
		if c == '"' {
			value := me.forString()
			token := valueToken("string", value)
			fmt.Println(len(tokens), token.string())
			tokens = append(tokens, token)
			continue
		}
		if strings.IndexByte("+-*/()=.", c) >= 0 {
			stream.next()
			token := simpleToken(string(c))
			fmt.Println(len(tokens), token.string())
			tokens = append(tokens, token)
			continue
		}
		if c == '\n' {
			stream.next()
			token := simpleToken("line")
			fmt.Println(len(tokens), token.string())
			tokens = append(tokens, token)
			continue
		}
		panic("unknown token " + stream.fail())
	}
	token := simpleToken("eof")
	fmt.Println(len(tokens), token.string())
	tokens = append(tokens, token)
	return tokens
}
