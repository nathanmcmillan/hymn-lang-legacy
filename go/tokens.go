package main

import (
	"fmt"
	"strings"
)

var keywords = map[string]bool{
	"function": true,
	"return":   true,
	"class":    true,
	"new":      true,
	"true":     true,
	"false":    true,
	"free":     true,
	"if":       true,
	"elif":     true,
	"else":     true,
	"for":      true,
	"continue": true,
	"break":    true,
}

func (me *token) string() string {
	if me.value == "" {
		return fmt.Sprintf("{depth:%d, type:%s}", me.depth, me.is)
	}
	return fmt.Sprintf("{depth:%d, type:%s, value:%s}", me.depth, me.is, me.value)
}

func simpleToken(depth int, is string) *token {
	t := &token{}
	t.depth = depth
	t.is = is
	t.value = ""
	return t
}

func valueToken(depth int, is, value string) *token {
	t := &token{}
	t.depth = depth
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
	depth := 0
	updateDepth := true
	for stream.pos < size {
		space := me.forSpace()
		if updateDepth {
			if space%2 != 0 {
				panic("bad spacing" + stream.fail())
			}
			depth = space / 2
			updateDepth = false
		}
		if stream.pos == size {
			break
		}
		typed, number := me.forNumber()
		if number != "" {
			token := valueToken(depth, typed, number)
			tokens = append(tokens, token)
			continue
		}
		word := me.forWord()
		if word != "" {
			var token *token
			if _, ok := keywords[word]; ok {
				if word == "true" || word == "false" {
					token = valueToken(depth, "bool", word)
				} else {
					token = simpleToken(depth, word)
				}
			} else {
				token = valueToken(depth, "id", word)
			}
			tokens = append(tokens, token)
			continue
		}
		c := stream.peek()
		if c == '"' {
			value := me.forString()
			token := valueToken(depth, "string", value)
			tokens = append(tokens, token)
			continue
		}
		if c == '-' {
			stream.next()
			peek := stream.peek()
			var token *token
			if peek == '>' {
				stream.next()
				token = simpleToken(depth, "return-type")
			} else if peek == '=' {
				stream.next()
				token = simpleToken(depth, "-=")
			} else {
				token = simpleToken(depth, "-")
			}
			tokens = append(tokens, token)
			continue
		}
		if strings.IndexByte("+*/", c) >= 0 {
			stream.next()
			op := string(c)
			peek := stream.peek()
			if peek == '=' {
				stream.next()
				op += "="
			}
			token := simpleToken(depth, op)
			tokens = append(tokens, token)
			continue
		}
		if c == '<' || c == '>' {
			stream.next()
			var token *token
			if stream.peek() == '=' {
				stream.next()
				token = simpleToken(depth, string(c)+"=")
			} else {
				token = simpleToken(depth, string(c))
			}
			tokens = append(tokens, token)
			continue
		}
		if strings.IndexByte(";()=.:[]", c) >= 0 {
			stream.next()
			token := simpleToken(depth, string(c))
			tokens = append(tokens, token)
			continue
		}
		if c == '\n' {
			stream.next()
			token := simpleToken(0, "line")
			tokens = append(tokens, token)
			updateDepth = true
			continue
		}
		panic("unknown token " + stream.fail())
	}
	token := simpleToken(0, "eof")
	tokens = append(tokens, token)
	return tokens
}
