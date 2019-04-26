package main

import (
	"fmt"
	"strings"
)

type token struct {
	t string
	v string
}

func (me *token) string() string {
	if me.v == "" {
		return fmt.Sprintf("{type:%s}", me.t)
	}
	return fmt.Sprintf("{type:%s, value:%s}", me.t, me.v)
}

const syntaxSpace = ' '
const syntaxComment = '#'
const syntaxNewLine = '\n'
const syntaxParenStart = '('
const syntaxParenEnd = ')'
const syntaxQuote = '"'
const syntaxAdd = '+'

const tokenSet = "set"
const tokenConst = "constant"
const tokenFn = "function"
const tokenEnd = "end"
const tokenParenStart = "parenStart"
const tokenParenEnd = "parenEnd"
const tokenID = "id"
const tokenUnknown = "?"
const tokenNewLine = "line"
const tokenNumber = "number"
const tokenString = "string"
const tokenAdd = "add"
const tokenSubtract = "subtract"
const tokenGreater = ">"
const tokenGreaterOrEq = ">="

var keywords = map[string]bool{
	"def":       true,
	"and":       true,
	"or":        true,
	"if":        true,
	"elif":      true,
	"else":      true,
	"return":    true,
	"int":       true,
	"int32":     true,
	"int64":     true,
	"float32":   true,
	"float64":   true,
	"bool":      true,
	"string":    true,
	"object":    true,
	"interface": true,
	"global":    true,
	"import":    true,
	"module":    true,
	"mutable":   true,
	"immutable": true,
	"maybe":     true,
	"none":      true,
}

var syntaxToToken = map[string]string{
	"set":   tokenSet,
	"const": tokenConst,
	"def":   tokenFn,
	"(":     tokenParenStart,
	")":     tokenParenEnd,
	"\n":    tokenNewLine,
	"+":     tokenAdd,
	"-":     tokenSubtract,
	"--":    tokenEnd,
}

func newToken(typeOf, valueOf string) *token {
	t := &token{}
	t.t = typeOf
	t.v = valueOf
	return t
}

type tokenizer struct {
	stream  *stream
	current string
}

func tokenize(stream *stream) []*token {
	tokenizer := tokenizer{}
	tokenizer.stream = stream
	tokens := make([]*token, 0)
	num := len(stream.data)
	for stream.pos < num {
		token := tokenizer.next()
		if token != nil {
			fmt.Println("token:", token.string())
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func (me *tokenizer) next() *token {
	stream := me.stream
	me.skipSpace()
	if stream.eof() {
		return newToken(tokenNewLine, "")
	}
	c := stream.peek()
	if c == syntaxComment {
		me.skipComment()
		return newToken(tokenNewLine, "")
	} else if c == syntaxNewLine {
		stream.next()
		return newToken(tokenNewLine, "")
	} else if c == syntaxParenStart {
		stream.next()
		return newToken(tokenParenStart, "")
	} else if c == syntaxParenEnd {
		stream.next()
		return newToken(tokenParenEnd, "")
	} else if c == syntaxQuote {
		return me.readString()
	} else if c == '-' {
		return me.readMinus()
	} else if digit(c) {
		return me.readNumber()
	} else if letter(c) {
		return me.readWord()
	}
	panic("tokenize failed on " + stream.fail())
}

func digit(c byte) bool {
	return strings.IndexByte("0123456789", c) >= 0
}

func letter(c byte) bool {
	return strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c) >= 0
}

func (me *tokenizer) readMinus() *token {
	stream := me.stream
	stream.next()
	c := stream.peek()
	if c == '-' {
		stream.next()
		return newToken(tokenEnd, "")
	}
	return newToken(tokenSubtract, "")
}

func (me *tokenizer) readString() *token {
	stream := me.stream
	stream.next()
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.next()
		if c == syntaxQuote {
			break
		}
		value.WriteByte(c)
	}
	return newToken(tokenString, value.String())
}

func (me *tokenizer) readNumber() *token {
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
	return newToken(tokenNumber, value.String())
}

func (me *tokenizer) readWord() *token {
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
	word := value.String()
	if _, ok := keywords[word]; ok {
		return newToken(word, "")
	}
	return newToken(tokenID, value.String())
}

func (me *tokenizer) skipSpace() {
	stream := me.stream
	for !stream.eof() {
		c := stream.peek()
		if c != syntaxSpace {
			break
		}
		stream.next()
	}
}

func (me *tokenizer) skipComment() {
	stream := me.stream
	for !me.stream.eof() {
		c := stream.peek()
		if c != syntaxNewLine {
			break
		}
		stream.next()
	}
}
