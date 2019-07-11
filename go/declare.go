package main

import "strings"

func (me *parser) declareType() string {
	typed := ""
	if me.token.is == "[" {
		me.eat("[")
		me.eat("]")
		typed += "[]"
	}

	value := me.token.value
	me.eat("id")
	typed += value

	if _, ok := me.hmfile.imports[value]; ok {
		me.eat(".")
		typed += "."
		value = me.token.value
		me.eat("id")
		typed += value
	}

	if me.token.is == "<" {
		me.eat("<")
		typed += "<"
		ix := 0
		for {
			if ix > 0 {
				typed += "," + me.token.value
			} else {
				typed += me.token.value
			}
			me.eat("id")
			if me.token.is == "delim" {
				me.eat("delim")
				ix++
				continue
			}
			if me.token.is == ">" {
				break
			}
			panic(me.fail() + "bad token \"" + me.token.is + "\" in generic type declaration")
		}
		me.eat(">")
		typed += ">"
	}

	return typed
}

func (me *parser) nameOfClassFunc(classname, funcname string) string {
	return classname + "_" + funcname
}

func typeOfArray(typed string) string {
	return typed[2:]
}

func checkIsArray(typed string) bool {
	return strings.HasPrefix(typed, "[]")
}

func (me *parser) assignable(n *node) bool {
	return n.is == "variable" || n.is == "member-variable" || n.is == "array-member"
}
