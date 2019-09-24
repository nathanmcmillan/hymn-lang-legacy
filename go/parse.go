package main

import (
	"fmt"
	"os"
)

type parser struct {
	hmfile *hmfile
	tokens *tokenizer
	token  *token
	pos    int
	line   int
	file   *os.File
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

	var tokenFile *os.File
	var treeFile *os.File

	if debug {
		fmt.Println("=== " + name + " parse ===")

		fileTokens := out + "/" + name + ".tokens"
		if exists(fileTokens) {
			os.Remove(fileTokens)
		}
		tokenFile = create(fileTokens)
		defer tokenFile.Close()

		fileTree := out + "/" + name + ".tree"
		if exists(fileTree) {
			os.Remove(fileTree)
		}
		treeFile = create(fileTree)
		defer treeFile.Close()
	}

	stream := newStream(data)
	parsing := parser{}
	parsing.hmfile = me
	parsing.line = 1
	parsing.tokens = tokenize(stream, tokenFile)
	parsing.token = parsing.tokens.get(0)
	parsing.file = treeFile

	parsing.skipLines()
	for parsing.token.is != "eof" {
		parsing.fileExpression()
		if parsing.token.is == "line" {
			parsing.eat("line")
		}
	}

	treeFile.Truncate(0)
	treeFile.Seek(0, 0)
	treeFile.WriteString(me.dump())
	treeFile.WriteString("\n")
}

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

func (me *parser) replace(want, is string) {
	me.verify(want)
	me.token.is = is
}

func (me *parser) wordOrPrimitive() {
	me.verifyWordOrPrimitive()
	me.next()
}

func (me *parser) verifyWordOrPrimitive() {
	t := me.token.is
	if t == "id" {
		me.verify("id")
		return
	} else if _, ok := primitives[t]; ok {
		me.verify(t)
		return
	}
	me.verify("id or primitive")
}
