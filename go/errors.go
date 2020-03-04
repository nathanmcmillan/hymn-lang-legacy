package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type parseLine struct {
	number  int
	content string
}

type parseError struct {
	code        int
	description string
	module      *hmfile
	lines       []*parseLine
	line        int
	begin       int
	end         int
	report      string
	trace       string
}

func erc(parser *parser, code int) *parseError {
	return err(parser, code, "")
}

func err(parser *parser, code int, description string) *parseError {
	bytes := make([]byte, 1<<16)
	runtime.Stack(bytes, true)
	stacktrace := fmt.Sprintf("%s", bytes)

	lines := make([]*parseLine, 0)

	stream := parser.tokens.stream
	content := stream.data
	pos := stream.pos
	number := stream.line
	size := len(content)
	for pos >= size {
		pos--
	}

gather:
	for i := 0; i < 5; i++ {
		line := &strings.Builder{}
		for true {
			if pos == -1 {
				break gather
			}
			b := content[pos]
			if b == '\n' {
				number--
				break
			}
			line.WriteByte(b)
			pos--
		}
		lines = append(lines, &parseLine{number, line.String()})
		line.Reset()
	}

	pos = stream.pos
	b := &strings.Builder{}
	i := 0
	line := 0
	for {
		str := &strings.Builder{}
		for i < size {
			c := content[i]
			i++
			str.WriteByte(c)
			if c == '\n' {
				line++
				break
			}
		}
		if str.Len() == 0 {
			break
		}
		b.WriteString(fmt.Sprintf("%d: %s", line, str.String()))
		b.WriteString("")
	}
	report := b.String()

	e := &parseError{}
	e.code = code
	e.description = parser.fail() + description
	e.trace = stacktrace
	e.lines = lines
	e.report = report

	return e
}

func (me *parseError) print() string {
	out := ""
	for _, line := range me.lines {
		out += strconv.Itoa(line.number) + " |     " + line.content + "\n"
	}
	out += "--------------------------------------------------------------------------------\n"
	out += me.report
	out += "--------------------------------------------------------------------------------\n"
	out += fmt.Sprintf("Code: %04d%s", me.code, me.description)
	out += "\n--------------------------------------------------------------------------------\n"
	out += me.trace

	return out
}

func (me *parseError) simple() string {
	out := fmt.Sprintf("Code: %04d\n", me.code)
	return out
}
