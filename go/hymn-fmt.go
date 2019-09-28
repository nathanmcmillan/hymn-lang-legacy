package main

import (
	"fmt"
	"os"
	"strings"
)

func hymnFmt(path string) {
	in := read(path)

	stream := newStream(in)
	tokens := tokenize(stream, nil)
	pos := 0

	skipPrevious := make(map[string]bool)
	skipPrevious["line"] = true
	skipPrevious["("] = true
	skipPrevious["<"] = true
	skipPrevious["."] = true

	skipCurrent := make(map[string]bool)
	skipCurrent["line"] = true
	skipCurrent["("] = true
	skipCurrent[")"] = true
	skipCurrent["<"] = true
	skipCurrent[">"] = true
	skipCurrent["."] = true
	skipCurrent[","] = true

	var out strings.Builder
	var previous *token

	newline := false

	for {
		token := tokens.get(pos)
		if token.is == "eof" {
			break
		}

		if newline {
			out.WriteString("\n")
			out.WriteString(hymnNewLine(token.depth))
			newline = false
		} else {
			if previous == nil {
			} else if _, ok := skipPrevious[previous.is]; ok {
			} else if _, ok := skipCurrent[token.is]; ok {
			} else {
				out.WriteString(" ")
			}
		}

		if token.value == "" {
			if token.is == "line" {
				newline = true
			} else {
				out.WriteString(hymnFmtToken(token))
			}
		} else {
			out.WriteString(hymnFmtToken(token))
		}

		pos++
		previous = token
	}

	content := string(out.String())
	fmt.Println(content)

	n := path + ".fmt"
	f := create(n)
	f.WriteString(content)
	f.Close()

	os.Remove(path)
	os.Rename(n, path)
}

func hymnNewLine(depth int) string {
	space := ""
	for i := 0; i < depth; i++ {
		space += "\t"
	}
	return space
}

func hymnFmtToken(t *token) string {
	if t.value == "" {
		return t.is
	}
	if t.is == TokenStringLiteral {
		return "\"" + t.value + "\""
	}
	return t.value
}
