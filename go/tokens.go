package main

import (
	"fmt"
	"os"
	"strings"
)

// tokens
const (
	TokenIntLiteral     = "int-v"
	TokenFloatLiteral   = "float-v"
	TokenStringLiteral  = "string-v"
	TokenCharLiteral    = "char-v"
	TokenBooleanLiteral = "bool-v"
	TokenInt            = "int"
	TokenInt8           = "int8"
	TokenInt16          = "int16"
	TokenInt32          = "int32"
	TokenInt64          = "int64"
	TokenUInt           = "uint"
	TokenUInt8          = "uint8"
	TokenUInt16         = "uint16"
	TokenUInt32         = "uint32"
	TokenUInt64         = "uint64"
	TokenFloat          = "float"
	TokenFloat32        = "float32"
	TokenFloat64        = "float64"
	TokenString         = "string"
	TokenRawString      = "string-raw"
	TokenChar           = "char"
	TokenBoolean        = "bool"
)

var keywords = map[string]bool{
	"import":     true,
	"macro":      true,
	"return":     true,
	"true":       true,
	"false":      true,
	"free":       true,
	"not":        true,
	"if":         true,
	"elif":       true,
	"else":       true,
	"for":        true,
	"while":      true,
	"continue":   true,
	"break":      true,
	"try":        true,
	"catch":      true,
	"mutable":    true,
	"static":     true,
	"and":        true,
	"or":         true,
	"as":         true,
	"enum":       true,
	"match":      true,
	"panic":      true,
	"pass":       true,
	"none":       true,
	"some":       true,
	"maybe":      true,
	"goto":       true,
	"label":      true,
	"async":      true,
	"yield":      true,
	"await":      true,
	"ifdef":      true,
	"ifndef":     true,
	"elsedef":    true,
	"enddef":     true,
	"defc":       true,
	"endc":       true,
	"alias":      true,
	"is":         true,
	"iterate":    true,
	"in":         true,
	"def":        true,
	"class":      true,
	"interface":  true,
	"implements": true,
	"with":       true,
	"where":      true,
	TokenInt:     true,
	TokenInt8:    true,
	TokenInt16:   true,
	TokenInt32:   true,
	TokenInt64:   true,
	TokenUInt:    true,
	TokenUInt8:   true,
	TokenUInt16:  true,
	TokenUInt32:  true,
	TokenUInt64:  true,
	TokenFloat:   true,
	TokenFloat32: true,
	TokenFloat64: true,
	TokenChar:    true,
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
		return fmt.Sprintf("{\"depth\": %d, \"type\": \"%s\"}", me.depth, me.is)
	}
	return fmt.Sprintf("{\"depth\": %d, \"type\": \"%s\", \"value\": \"%s\"}", me.depth, me.is, me.value)
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

func (me *tokenizer) forNumber() (string, string, *tokenizeError) {
	stream := me.stream
	typed := TokenIntLiteral
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.peek()
		if c == '.' {
			if value.Len() == 0 {
				break
			}
			typed = TokenFloatLiteral
			value.WriteByte(c)
			stream.next()
			if !digit(stream.peek()) {
				return "", "", me.exception("Digit must follow after dot. " + stream.fail())
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
	return typed, value.String(), nil
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

func (me *tokenizer) consumeDepth() *tokenizeError {
	spaces := me.depth * 4
	stream := me.stream
	for i := 0; i < spaces && !stream.eof(); i++ {
		c := stream.next()
		if c != ' ' {
			return me.exception("Bad spacing " + stream.fail())
		}
	}
	return nil
}

func (me *tokenizer) forString() (string, *tokenizeError) {
	stream := me.stream
	stream.next()
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.next()
		if c == '"' {
			break
		}
		if c == '\\' {
			peek := stream.peek()
			if peek == '\n' || peek == ' ' {
				if peek == ' ' {
					me.forSpace()
					peek = stream.peek()
					if peek != '\n' {
						return "", me.exception("Bad string " + stream.fail())
					}
				}
				stream.next()
				e := me.consumeDepth()
				if e != nil {
					return "", e
				}
				continue
			}
		}
		value.WriteByte(c)
	}
	return value.String(), nil
}

func (me *tokenizer) forLineComment() string {
	stream := me.stream
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.next()
		if c == '\n' {
			break
		}
		value.WriteByte(c)
	}
	return value.String()
}

func (me *tokenizer) push(t *token) {
	me.tokens = append(me.tokens, t)
	if me.file != nil {
		if len(me.tokens) > 1 {
			me.file.WriteString(",\n")
		}
		me.file.WriteString("        " + t.string())
	}
}

func (me *tokenizer) get(pos int) (*token, *tokenizeError) {
	if pos < len(me.tokens) {
		return me.tokens[pos], nil
	}
	stream := me.stream
	if stream.pos >= me.size {
		return me.eof, nil
	}
	space := me.forSpace()
	if me.updateDepth {
		if space%2 != 0 {
			return nil, me.exception("Bad spacing " + stream.fail())
		}
		me.depth = space / 4
		me.updateDepth = false
	}
	if stream.pos >= me.size {
		return me.eof, nil
	}
	typed, number, e := me.forNumber()
	if e != nil {
		return nil, e
	}
	if number != "" {
		token := me.valueToken(typed, number)
		me.push(token)
		return token, nil
	}
	word := me.forWord()
	if word != "" {
		var token *token
		if _, ok := keywords[word]; ok {
			if word == "true" || word == "false" {
				token = me.valueToken(TokenBooleanLiteral, word)
			} else if checkIsPrimitive(word) {
				token = me.valueToken(word, word)
			} else {
				token = me.simpleToken(word)
			}
		} else {
			token = me.valueToken("id", word)
		}
		me.push(token)
		return token, nil
	}
	c := stream.peek()
	if strings.IndexByte("$().[]_?,;", c) >= 0 {
		stream.next()
		token := me.simpleToken(string(c))
		me.push(token)
		return token, nil
	}
	if c == '\'' {
		stream.next()
		value := ""
		ischar := false
		if stream.peek() == '\\' {
			value += "\\"
			ischar = true
			stream.next()
		}
		peek := stream.doublePeek()
		if peek == '\'' {
			ischar = true
		} else if ischar {
			return nil, me.exception("Expecting character literal " + stream.fail())
		}
		if ischar {
			value += string(stream.peek())
			stream.next()
			stream.next()
			token := me.valueToken(TokenCharLiteral, "'"+value+"'")
			me.push(token)
			return token, nil
		}
		token := me.simpleToken(string(c))
		me.push(token)
		return token, nil
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
		return token, nil
	}
	if c == '"' {
		value, e := me.forString()
		if e != nil {
			return nil, e
		}
		token := me.valueToken(TokenStringLiteral, value)
		me.push(token)
		return token, nil
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
		return token, nil
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
		return token, nil
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
		return token, nil
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
		return token, nil
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
		return token, nil
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
		return token, nil
	}
	if c == '\n' {
		stream.next()
		token := me.tokenFor(0, "line")
		me.push(token)
		me.updateDepth = true
		return token, nil
	}
	if c == '#' {
		stream.next()
		value := me.forLineComment()
		token := me.valueToken("comment", value)
		me.push(token)
		return token, nil
	}
	return nil, me.exception("Unknown token " + stream.fail())
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
