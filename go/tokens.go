package main

import (
	"fmt"
	"os"
	"strings"
)

var keywords = map[string]bool{
	"import":    true,
	"macro":     true,
	"return":    true,
	"type":      true,
	"true":      true,
	"false":     true,
	"free":      true,
	"not":       true,
	"if":        true,
	"elif":      true,
	"else":      true,
	"for":       true,
	"continue":  true,
	"break":     true,
	"mutable":   true,
	"immutable": true,
	"and":       true,
	"or":        true,
	"as":        true,
	"enum":      true,
	"match":     true,
	"panic":     true,
	"pass":      true,
	"none":      true,
	"some":      true,
	"maybe":     true,
	"goto":      true,
	"label":     true,
	"async":     true,
	"def":       true,
	"ifdef":     true,
	"ifndef":    true,
	"elsedef":   true,
	"enddef":    true,
	"defc":      true,
	"endc":      true,
	"alias":     true,
	"is":        true,
}

type token struct {
	depth int
	is    string
	value string
}

type tokenizer struct {
	stream      *stream
	current     string
	tokens      []*token
	eof         *token
	size        int
	depth       int
	updateDepth bool
	file        *os.File
}

func (me *token) string() string {
	if me.value == "" {
		return fmt.Sprintf("{depth:%d, type:%s}", me.depth, me.is)
	}
	return fmt.Sprintf("{depth:%d, type:%s, value:%s}", me.depth, me.is, me.value)
}

func digit(c byte) bool {
	return strings.IndexByte("0123456789", c) >= 0
}

func letter(c byte) bool {
	return strings.IndexByte("abcdefghijklmnopqrstuvwxyz", c) >= 0
}

func (me *tokenizer) forSpace() int {
	stream := me.stream
	count := 0
	for !stream.eof() {
		c := stream.peek()
		if c == ' ' {
			count++
			stream.next()
		} else if c == '\t' {
			count += 2
			stream.next()
		} else {
			break
		}
	}
	return count
}

func (me *tokenizer) simpleToken(is string) *token {
	return me.tokenFor(me.depth, is)
}

func (me *tokenizer) tokenFor(depth int, is string) *token {
	t := &token{}
	t.depth = depth
	t.is = is
	t.value = ""
	return t
}

func (me *tokenizer) valueToken(is, value string) *token {
	t := &token{}
	t.depth = me.depth
	t.is = is
	t.value = value
	return t
}

func (me *tokenizer) forNumber() (string, string) {
	stream := me.stream
	typed := "int"
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.peek()
		if c == '.' {
			if value.Len() == 0 {
				break
			}
			typed = "float"
			value.WriteByte(c)
			stream.next()
			if !digit(stream.peek()) {
				panic("digit must follow after dot. " + stream.fail())
			}
			continue
		}
		if digit(c) {
			value.WriteByte(c)
			stream.next()
			continue
		}
		break
	}
	return typed, value.String()
}

func (me *tokenizer) forWord() string {
	stream := me.stream
	value := &strings.Builder{}
	first := true
	for !stream.eof() {
		c := stream.peek()
		if !letter(c) {
			if first {
				break
			} else if !digit(c) && c != '_' {
				break
			}
		}
		value.WriteByte(c)
		stream.next()
		first = false
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

func (me *tokenizer) forComment() string {
	stream := me.stream
	value := &strings.Builder{}
	nest := 1
	for !stream.eof() {
		c := stream.next()
		if c == '(' {
			c2 := stream.peek()
			if c2 == '*' {
				nest++
			}
		}
		if c == '*' {
			c2 := stream.peek()
			if c2 == ')' {
				nest--
				if nest == 0 {
					stream.next()
					break
				}
			}
		}
		value.WriteByte(c)
	}
	return value.String()
}

func (me *tokenizer) push(t *token) {
	me.tokens = append(me.tokens, t)
	if me.file != nil {
		me.file.WriteString(t.string() + "\n")
	}
}

func (me *tokenizer) get(pos int) *token {
	if pos < len(me.tokens) {
		return me.tokens[pos]
	}
	stream := me.stream
	if stream.pos >= me.size {
		return me.eof
	}
	space := me.forSpace()
	if me.updateDepth {
		if space%2 != 0 {
			panic(stream.fail() + "bad spacing")
		}
		me.depth = space / 2
		me.updateDepth = false
	}
	if stream.pos >= me.size {
		return me.eof
	}
	typed, number := me.forNumber()
	if number != "" {
		token := me.valueToken(typed, number)
		me.push(token)
		return token
	}
	word := me.forWord()
	if word != "" {
		var token *token
		if _, ok := keywords[word]; ok {
			if word == "true" || word == "false" {
				token = me.valueToken("bool", word)
			} else {
				token = me.simpleToken(word)
			}
		} else {
			token = me.valueToken("id", word)
		}
		me.push(token)
		return token
	}
	c := stream.peek()
	if c == '(' {
		stream.next()
		peek := stream.peek()
		if peek == '*' {
			stream.next()
			// TODO buggy
			// value := me.forComment()
			// token := me.valueToken("comment", value)
			// me.push(token)
			// return token
			me.forComment()
			return me.get(pos)
		}
		token := me.simpleToken("(")
		me.push(token)
		return token
	}
	if strings.IndexByte("$).[]'_?", c) >= 0 {
		stream.next()
		token := me.simpleToken(string(c))
		me.push(token)
		return token
	}
	if c == ':' {
		stream.next()
		peek := stream.peek()
		var token *token
		if peek == '=' {
			stream.next()
			token = me.simpleToken(":=")
		} else {
			token = me.simpleToken(":")
		}
		me.push(token)
		return token
	}
	if c == '"' {
		value := me.forString()
		token := me.valueToken("string", value)
		me.push(token)
		return token
	}
	if c == '=' {
		stream.next()
		op := "="
		peek := stream.peek()
		if peek == '>' {
			stream.next()
			op = "=>"
		} else if peek == '=' {
			stream.next()
			op = "=="
		}
		token := me.simpleToken(op)
		me.push(token)
		return token
	}
	if c == '-' {
		stream.next()
		peek := stream.peek()
		var token *token
		if peek == '=' {
			stream.next()
			token = me.simpleToken("-=")
		} else if peek == '>' {
			stream.next()
			token = me.simpleToken("->")
		} else {
			token = me.simpleToken("-")
		}
		me.push(token)
		return token
	}
	if strings.IndexByte("+*/%&|^", c) >= 0 {
		stream.next()
		op := string(c)
		peek := stream.peek()
		if peek == '=' {
			stream.next()
			op += "="
		}
		token := me.simpleToken(op)
		me.push(token)
		return token
	}
	if c == '!' {
		stream.next()
		var token *token
		if stream.peek() == '=' {
			stream.next()
			token = me.simpleToken("!=")
		} else {
			token = me.simpleToken(string(c))
		}
		me.push(token)
		return token
	}
	if c == '>' {
		stream.next()
		var token *token
		if stream.peek() == '=' {
			stream.next()
			token = me.simpleToken(">=")
		} else if stream.peek() == '>' && stream.doublePeek() == '=' {
			stream.next()
			stream.next()
			token = me.simpleToken(">>=")
		} else {
			token = me.simpleToken(string(c))
		}
		me.push(token)
		return token
	}
	if c == '<' {
		stream.next()
		var token *token
		if stream.peek() == '=' {
			stream.next()
			token = me.simpleToken("<=")
		} else if stream.peek() == '<' {
			stream.next()
			if stream.peek() == '=' {
				stream.next()
				token = me.simpleToken("<<=")
			} else {
				token = me.simpleToken("<<")
			}
		} else {
			token = me.simpleToken(string(c))
		}
		me.push(token)
		return token
	}
	if c == ',' {
		stream.next()
		token := me.simpleToken("delim")
		me.push(token)
		return token
	}
	if c == '\n' {
		stream.next()
		token := me.tokenFor(0, "line")
		me.push(token)
		me.updateDepth = true
		return token
	}
	panic("unknown token " + stream.fail())
}

func tokenize(stream *stream, file *os.File) *tokenizer {
	me := &tokenizer{}
	me.stream = stream
	me.tokens = make([]*token, 0)
	me.eof = me.tokenFor(0, "eof")
	me.depth = 0
	me.updateDepth = true
	me.size = len(stream.data)
	me.file = file
	return me
}
