package main

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	printStacktrace = true
)

type parseLine struct {
	number  int
	content string
}

type parseError struct {
	code        int
	tokens      string
	name        string
	description string
	hint        string
	lines       []*parseLine
	begin       int
	end         int
	module      string
	report      string
	trace       string
}

func tokenToParseError(p *parser, t *tokenizeError) *parseError {
	return err(p, ECodeUnexpectedToken, t.reason)
}

func erc(parser *parser, code int) *parseError {
	return err(parser, code, "")
}

func err(parser *parser, code int, description string) *parseError {
	return errh(parser, code, description, "")
}

func errh(parser *parser, code int, description, hint string) *parseError {
	bytes := make([]byte, 1<<16)
	runtime.Stack(bytes, true)
	stacktrace := fmt.Sprintf("%s", bytes)

	e := &parseError{}
	e.code = code
	e.description = description
	if hint != "" {
		e.hint = "Hint: " + hint
	}
	e.trace = stacktrace

	report := ""
	if parser != nil {
		e.tokens = parser.fail()
		e.module = parser.hmfile.name

		stream := parser.tokens.stream
		content := stream.data
		number := stream.line
		size := len(content)

		i := 0
		line := 0
	find:
		for {
			str := &strings.Builder{}
			for i < size {
				c := content[i]
				i++
				str.WriteByte(c)
				if c == '\n' {
					if line == number {
						report = fmt.Sprintf("%d: %s", line, str.String())
						break find
					}
					line++
					break
				}
			}
			if str.Len() == 0 {
				break
			}
		}
	}
	e.report = report

	return e
}

func (me *parseError) print() string {
	out := "\n"
	printfile := ""
	if me.module != "" {
		printfile = me.module + ".hm"
	}
	top := fmt.Sprintf("-- Code: %04d ", me.code)
	for len(top)+len(printfile) != 79 {
		top += "-"
	}
	out += top + " " + printfile
	out += "\n\n"
	out += me.description
	if me.code == ECodeUnexpectedToken {
		out += "\n" + me.tokens
	}
	out += "\n\n"
	out += me.report
	if me.hint != "" {
		out += "\n\n" + me.hint
	}
	if printStacktrace {
		out += "\n\n--------------------------------------------------------------------------------\n"
		out += me.trace
	}
	out += "\n\n"
	return out
}

func (me *parseError) simple() string {
	out := fmt.Sprintf("Code: %04d\n", me.code)
	return out
}
