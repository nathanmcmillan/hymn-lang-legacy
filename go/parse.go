package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type parser struct {
	hmfile *hmfile
	tokens *tokenizer
	token  *token
	pos    int
	line   int
	file   *os.File
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
	str.WriteString(me.tokens.get(me.pos).string())

	fn := me.hmfile.scope.fn
	if fn != nil {
		str.WriteString("\nCurrent Function: ")
		str.WriteString(fn.module.reference(fn.getname()))
	}

	if me.hmfile.program.peekRemapStack() != "" {
		str.WriteString("\nFunction Implementation Stack: [")
		for _, r := range me.hmfile.program.remapStack {
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
		token := me.token
		if token.is != "line" {
			break
		}
		me.next()
	}
}

func (me *hmfile) parse(out, path string) *parseError {
	name := fileName(path)
	data := read(path)

	var tokenFile *os.File
	var treeFile *os.File

	if debug {
		os.MkdirAll(out, os.ModePerm)
		fmt.Println("=== parse: " + name + " ===")

		if debugTokens {
			fileTokens := out + "/" + name + "-tokens.json"
			if exists(fileTokens) {
				os.Remove(fileTokens)
			}
			tokenFile = create(fileTokens)
			tokenFile.WriteString("{\n\t\"tokens\": [\n")
			defer tokenFile.Close()
		}

		if debugTree {
			fileTree := out + "/" + name + "-tree.json"
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

	parsing.hmfile = me
	parsing.line = 1
	parsing.tokens = tokenize(stream, tokenFile)
	parsing.token = parsing.tokens.get(0)
	parsing.file = treeFile

	parsing.skipLines()
	for parsing.token.is != "eof" {
		if er := parsing.fileExpression(); er != nil {
			return er
		}
		if parsing.token.is == "line" {
			if er := parsing.eat("line"); er != nil {
				return er
			}
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
	me.token = me.tokens.get(me.pos)
	if me.token.is == "line" || me.token.is == "comment" {
		me.line++
	}
}

func (me *parser) peek() *token {
	return me.tokens.get(me.pos + 1)
}

func (me *parser) verify(want string) *parseError {
	token := me.token
	if token.is != want {
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
	if t == "id" {
		return me.verify("id")
	} else if checkIsPrimitive(t) {
		return me.verify(t)
	}
	return me.verify("id or primitive")
}

func (me *parser) save() *parsepoint {
	return &parsepoint{pos: me.pos, line: me.line}
}

func (me *parser) jump(p *parsepoint) {
	me.pos = p.pos
	me.token = me.tokens.get(me.pos)
	me.line = p.line
}
