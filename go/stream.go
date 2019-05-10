package main

import (
	"fmt"
	"strings"
)

type stream struct {
	data []byte
	pos  int
	line int
	col  int
}

func newStream(data []byte) *stream {
	s := &stream{}
	s.data = data
	s.line = 1
	return s
}

func (me *stream) next() byte {
	c := me.data[me.pos]
	me.pos++
	if c == '\n' {
		me.line++
		me.col = 0
	} else {
		me.col++
	}
	return c
}

func (me *stream) peek() byte {
	return me.data[me.pos]
}

func (me *stream) eof() bool {
	return me.pos == len(me.data)
}

func (me *stream) fail() string {
	data := me.data
	c := data[me.pos]
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("line: %d column: %d char:[%c] for:\n", me.line, me.col, c))
	lenOfData := len(data)
	i := 0
	line := 0
	for {
		content := &strings.Builder{}
		for i < lenOfData {
			c := data[i]
			i++
			content.WriteByte(c)
			if c == '\n' {
				line++
				break
			}
		}
		if content.Len() == 0 {
			break
		}
		b.WriteString(fmt.Sprintf("%d: %s", line, content.String()))
	}
	return b.String()
}
