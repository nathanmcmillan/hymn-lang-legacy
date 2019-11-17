package main

import (
	"strings"
)

func fmtptr(ptr string) string {
	if strings.HasSuffix(ptr, "*") {
		return ptr + "*"
	}
	return ptr + " *"
}

func fmtassignspace(s string) string {
	if strings.HasSuffix(s, "*") {
		return s
	}
	return s + " "
}

func (me *cfile) maybeLet(code string, attributes map[string]string) string {
	if code == "" || strings.HasPrefix(code, "[") {
		return ""
	}
	if _, ok := attributes["stack"]; ok {
		return ""
	}
	return " = "
}

func (me *cfile) maybeColon(code string) string {
	size := len(code)
	if size == 0 {
		return ""
	}
	last := code[size-1]
	if last == '}' || last == ':' || last == ';' || last == '\n' {
		return ""
	}
	return ";"
}

func (me *cfile) maybeNewLine(code string) string {
	if code[len(code)-1] == '\n' {
		return ""
	}
	return "\n"
}

func (me *cfile) maybeFmc(code string, depth int) string {
	if code == "" || code[0] == spaceChar {
		return ""
	}
	return fmc(depth)
}
