package main

import (
	"fmt"
	"os"
)

func (me *parser) next() {
	me.pos++
	me.token = me.tokens.get(me.pos)
	if me.token.is == "line" {
		me.line++
	}
}

func (me *parser) peek() *token {
	return me.tokens.get(me.pos + 1)
}

func (me *parser) fail() string {
	return fmt.Sprintf("line %d, token %s\n", me.line, me.tokens.get(me.pos).string())
}

func (me *parser) skipLines() {
	for me.token.is != "eof" {
		token := me.token
		if token.is != "line" {
			break
		}
		me.next()
	}
}

func (me *hmfile) parse(out, path string) {
	name := fileName(path)
	data := read(path)
	if debug {
		fmt.Println("=== " + name + " parse ===")
	}
	stream := newStream(data)
	parsing := parser{}
	parsing.hmfile = me
	parsing.line = 1
	parsing.tokens = tokenize(stream)
	parsing.token = parsing.tokens.get(0)
	parsing.skipLines()
	for parsing.token.is != "eof" {
		parsing.fileExpression()
		if parsing.token.is == "line" {
			parsing.eat("line")
		}
	}

	if debug {
		dump := ""
		for _, token := range parsing.tokens.tokens {
			dump += token.string() + "\n"
		}
		fileTokens := out + "/" + name + ".tokens"
		if exists(fileTokens) {
			os.Remove(fileTokens)
		}
		create(fileTokens, dump)
	}

	delete(me.functions, "echo")
}

func (me *parser) verify(want string) {
	token := me.token
	if token.is != want {
		panic(me.fail() + "unexpected token was " + token.string() + " instead of {type:" + want + "}")
	}
}

func (me *parser) eat(want string) {
	me.verify(want)
	me.next()
}
