package main

import "fmt"

type parser struct {
	tokens []*token
	pos    int
}

type instruction struct {
	d string
}

type program struct {
	multi []string
}

func (me *instruction) string() string {
	return me.d
}

func parse(tokens []*token) []*instruction {
	parser := parser{}
	parser.tokens = tokens
	instructions := make([]*instruction, 0)
	num := len(tokens)
	for parser.pos < num {
		instruction := parser.eval()
		if instruction != nil {
			fmt.Println("instructions:", instruction.string())
			instructions = append(instructions, instruction)
		}
	}
	return instructions
}

func (me *parser) eval() *instruction {
	me.skipLines()
	if me.eot() {
		return nil
	}
	token := me.peek()
	if token.is == tokenFunc {
		return me.parseFunc()
	}
	panic("parser failed on " + me.fail())
}

func (me *parser) parseFunc() *instruction {
	me.next()
	i := &instruction{}
	if me.eot() {
		panic("failed to parse function")
	}
	return i
}

func (me *parser) skipLines() {
	for !me.eot() {
		token := me.peek()
		if token.is != tokenLine {
			break
		}
		me.next()
	}
}

func (me *parser) next() *token {
	t := me.tokens[me.pos]
	me.pos++
	return t
}

func (me *parser) peek() *token {
	return me.tokens[me.pos]
}

func (me *parser) eot() bool {
	return me.pos == len(me.tokens)
}

func (me *parser) fail() string {
	return fmt.Sprintf("token: %d %s\n", me.pos, me.tokens[me.pos].string())
}
