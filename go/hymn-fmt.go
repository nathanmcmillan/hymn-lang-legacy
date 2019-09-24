package main

import (
	"fmt"
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

	skipCombo := make(map[string]map[string]bool)
	// skipCombo["("] = make(map[string]bool)
	// skipCombo["("][TokenInt] = true
	// skipCombo["("]["id"] = true
	// skipCombo["id"] = make(map[string]bool)
	// skipCombo["id"]["("] = true
	// skipCombo["id"][")"] = true
	// skipCombo[")"] = make(map[string]bool)

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
			} else if _, ok := skipCombo[previous.is][token.is]; ok {
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

	fmt.Println(string(out.String()))
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
	if t.is == TokenString {
		return "\"" + t.value + "\""
	}
	return t.value
}
