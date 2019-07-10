package main

import (
	"fmt"
	"strings"
)

var keywords = map[string]bool{
	"import":    true,
	"macro":     true,
	"return":    true,
	"class":     true,
	"true":      true,
	"false":     true,
	"free":      true,
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
	return strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c) >= 0
}

func (me *tokenizer) forSpace() int {
	stream := me.stream
	count := 0
	for !stream.eof() {
		c := stream.peek()
		if c != ' ' || c == '\t' {
			break
		}
		count++
		stream.next()
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
			} else if !digit(c) {
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

func (me *tokenizer) forComment() {
	stream := me.stream
	stream.next()
	for !stream.eof() {
		c := stream.next()
		if c == '\n' {
			break
		}
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
			panic("bad spacing" + stream.fail())
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
		me.tokens = append(me.tokens, token)
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
		me.tokens = append(me.tokens, token)
		return token
	}
	c := stream.peek()
	if strings.IndexByte("().[]_", c) >= 0 {
		stream.next()
		token := me.simpleToken(string(c))
		me.tokens = append(me.tokens, token)
		return token
	}
	if c == '"' {
		value := me.forString()
		token := me.valueToken("string", value)
		me.tokens = append(me.tokens, token)
		return token
	}
	if c == '=' {
		stream.next()
		op := string(c)
		peek := stream.peek()
		if peek == '>' {
			stream.next()
			op += ">"
		}
		token := me.simpleToken(op)
		me.tokens = append(me.tokens, token)
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
		me.tokens = append(me.tokens, token)
		return token
	}
	if strings.IndexByte("+*/", c) >= 0 {
		stream.next()
		op := string(c)
		peek := stream.peek()
		if peek == '=' {
			stream.next()
			op += "="
		}
		token := me.simpleToken(op)
		me.tokens = append(me.tokens, token)
		return token
	}
	if c == '<' || c == '>' || c == '!' {
		stream.next()
		var token *token
		if stream.peek() == '=' {
			stream.next()
			token = me.simpleToken(string(c) + "=")
		} else {
			token = me.simpleToken(string(c))
		}
		me.tokens = append(me.tokens, token)
		return token
	}
	if c == ',' {
		stream.next()
		token := me.simpleToken("delim")
		me.tokens = append(me.tokens, token)
		return token
	}
	if c == '#' {
		me.forComment()
		return me.get(pos)
	}
	if c == '\n' {
		stream.next()
		token := me.tokenFor(0, "line")
		me.tokens = append(me.tokens, token)
		me.updateDepth = true
		return token
	}
	panic("unknown token " + stream.fail())
}

func tokenize(stream *stream) *tokenizer {
	me := &tokenizer{}
	me.stream = stream
	me.tokens = make([]*token, 0)
	me.eof = me.tokenFor(0, "eof")
	me.depth = 0
	me.updateDepth = true
	me.size = len(stream.data)
	return me
}
