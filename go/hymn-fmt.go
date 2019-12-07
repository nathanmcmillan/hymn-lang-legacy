package main

import (
	"fmt"
	"os"
	"strings"
)

func execFormat(path string) {
	in := read(path)

	stream := newStream(in)
	tokens := tokenize(stream, nil)
	pos := 0

	skipWhitelist := make(map[string]map[string]bool)
	skipWhitelist["id"] = make(map[string]bool)
	skipWhitelist["id"]["["] = true
	skipWhitelist["["] = make(map[string]bool)
	skipWhitelist["["]["]"] = true

	// skipBlacklist := make(map[string]map[string]bool)
	// skipBlacklist["]"] = make(map[string]bool)
	// skipBlacklist["]"]["asd"] = true

	skipPrevious := make(map[string]bool)
	skipPrevious["line"] = true
	skipPrevious["("] = true
	skipPrevious["<"] = true
	skipPrevious["."] = true
	// skipPrevious["["] = true

	skipCurrent := make(map[string]bool)
	skipCurrent["line"] = true
	skipCurrent["("] = true
	skipCurrent[")"] = true
	skipCurrent["<"] = true
	skipCurrent[">"] = true
	skipCurrent["."] = true
	skipCurrent[","] = true
	// skipCurrent["["] = true

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
			ok := false
			if previous == nil {
				ok = true
			} else if _, ok = skipPrevious[previous.is]; ok {
			} else if _, ok = skipCurrent[token.is]; ok {
			} else {
				if _, ok2 := skipWhitelist[previous.is]; ok2 {
					_, ok = skipWhitelist[token.is]
				}
				// if _, ok2 := skipBlacklist[previous.is]; ok2 {
				// 	_, ok = skipBlacklist[token.is]
				// }
			}
			if !ok {
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
	if t.is == "comment" {
		return "(*" + t.value + "*)"
	}
	return t.value
}
