package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type parser struct {
	program *program
	hmfile  *hmfile
	tokens  *tokenizer
	token   *token
	pos     int
	line    int
	file    *os.File
}

type parsepoint struct {
	pos  int
	line int
}

func (me *parser) fail() string {
	var str strings.Builder
	str.WriteString("\nModule: ")
	str.WriteString(me.hmfile.name)
	str.WriteString("\nLine: ")
	str.WriteString(strconv.Itoa(me.line))
	str.WriteString("\nToken: ")
	t, e := me.tokens.get(me.pos)
	if e != nil {
		panic("Token error: " + e.reason)
	}
	str.WriteString(t.string())

	fn := me.hmfile.getFuncScope().fn
	if fn != nil {
		str.WriteString("\nCurrent Function: ")
		str.WriteString(fn.module.reference(fn.getname()))
	}

	if me.program.peekRemapStack() != "" {
		str.WriteString("\nFunction Implementation Stack: [")
		for _, r := range me.program.remapStack {
			str.WriteString("\n    ")
			str.WriteString(r)
		}
		str.WriteString("\n]")
	}

	str.WriteString("\nError: ")
	return str.String()
}

func (me *parser) skipLines() {
	for me.token.is != "eof" {
		if me.token.is != "line" {
			break
		}
		me.next()
	}
}

func (me *hmfile) parse() *parseError {

	source := me.path
	destination := me.destination

	name := fileName(source)
	data, er := read(source)
	if er != nil {
		return err(me.parser, ECodeSystemError, er.Error())
	}

	var tokenFile *os.File
	var treeFile *os.File

	if debug {
		os.MkdirAll(destination, os.ModePerm)
		fmt.Println("parse>", name)

		if debugTokens {
			fileTokens := destination + "/" + name + "-tokens.json"
			if exists(fileTokens) {
				os.Remove(fileTokens)
			}
			tokenFile = create(fileTokens)
			tokenFile.WriteString("{\n\t\"tokens\": [\n")
			defer tokenFile.Close()
		}

		if debugTree {
			fileTree := destination + "/" + name + "-tree.json"
			if exists(fileTree) {
				os.Remove(fileTree)
			}
			treeFile = create(fileTree)
			defer treeFile.Close()
		}
	}

	stream := newStream(data)

	parsing := &parser{}
	me.parser = parsing

	parsing.program = me.program
	parsing.hmfile = me
	parsing.line = 1
	parsing.tokens = tokenize(stream, tokenFile)
	var e *tokenizeError
	parsing.token, e = parsing.tokens.get(0)
	if e != nil {
		return tokenToParseError(parsing, e)
	}
	parsing.file = treeFile

	parsing.skipLines()
	for parsing.token.is != "eof" {
		if er := parsing.statement(); er != nil {
			return er
		}
		if parsing.isNewLine() {
			parsing.newLine()
		}
	}

	if er := parsing.verifyFile(); er != nil {
		return er
	}

	if tokenFile != nil {
		tokenFile.WriteString("\n\t]\n}\n")
	}

	if treeFile != nil {
		treeFile.Truncate(0)
		treeFile.Seek(0, 0)
		treeFile.WriteString(me.string())
		treeFile.WriteString("\n")
	}

	return nil
}

func (me *parser) next() {
	me.pos++
	var e *tokenizeError
	me.token, e = me.tokens.get(me.pos)
	if e != nil {
		panic("Token error: " + e.reason)
	}
	if me.token.is == "line" || me.token.is == "comment" {
		me.line++
	}
}

func (me *parser) peek() *token {
	t, e := me.tokens.get(me.pos + 1)
	if e != nil {
		panic("Token error: " + e.reason)
	}
	return t
}

func (me *parser) doublePeek() *token {
	t, e := me.tokens.get(me.pos + 2)
	if e != nil {
		panic("Token error: " + e.reason)
	}
	return t
}

func (me *parser) verify(want string) *parseError {
	token := me.token
	if token.is != want {
		if want == "line" && token.is == "comment" {
			return nil
		}
		return err(me, ECodeUnexpectedToken, "unexpected token was "+token.string()+" instead of {type:"+want+"}")
	}
	return nil
}

func (me *parser) eat(want string) *parseError {
	if er := me.verify(want); er != nil {
		return er
	}
	me.next()
	return nil
}

func (me *parser) replace(want, is string) *parseError {
	if er := me.verify(want); er != nil {
		return er
	}
	me.token.is = is
	return nil
}

func (me *parser) wordOrPrimitive() *parseError {
	if er := me.verifyWordOrPrimitive(); er != nil {
		return er
	}
	me.next()
	return nil
}

func (me *parser) verifyWordOrPrimitive() *parseError {
	t := me.token.is
	if t == "id" || checkIsPrimitive(t) {
		return nil
	}
	return me.verify("id or primitive")
}

func (me *parser) newLine() *parseError {
	t := me.token.is
	if t == "line" {
		me.next()
		return nil
	} else if t == "comment" && me.peek().is == "line" {
		me.next()
		me.next()
		return nil
	}
	return me.verify("line")
}

func (me *parser) verifyNewLine() *parseError {
	if me.isNewLine() {
		return nil
	}
	return me.verify("line")
}

func (me *parser) isNewLine() bool {
	t := me.token.is
	return t == "line" || (t == "comment" && me.peek().is == "line")
}

func (me *parser) save() *parsepoint {
	return &parsepoint{pos: me.pos, line: me.line}
}

func (me *parser) jump(p *parsepoint) {
	var e *tokenizeError
	me.pos = p.pos
	me.token, e = me.tokens.get(me.pos)
	if e != nil {
		panic("Token error: " + e.reason)
	}
	me.line = p.line
}
