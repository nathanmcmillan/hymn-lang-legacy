package main

import "fmt"

func (me *parser) def() *node {
	me.eat("def")
	name := me.token.value
	me.eat("id")
	var value *node
	if me.token.is != "line" {
		value = me.calc(0, nil)
		fmt.Println("NEW DEF IS", name, ":=", value.string(0))
	} else {
		fmt.Println("NEW DEF IS", name)
	}
	me.eat("line")
	me.hmfile.defs[name] = value
	me.hmfile.namespace[name] = "def"
	me.hmfile.types[name] = ""
	return nil
}

func (me *parser) ifdef() *node {
	me.eat("ifdef")
	name := me.token.value
	me.eat("id")
	if _, ok := me.hmfile.defs[name]; ok {
		for {
			if me.token.is == "elsedef" || me.token.is == "enddef" {
				break
			}
			if me.token.is == "eof" {
				panic(me.fail() + "ifdef " + name + " missing enddef")
			}
		}
	} else {
		for {
			if me.token.is == "elsedef" || me.token.is == "enddef" {
				break
			}
			if me.token.is == "eof" {
				panic(me.fail() + "ifdef " + name + " missing enddef")
			}
		}
	}
	return nil
}

func (me *parser) elsedef() *node {
	me.eat("elsedef")
	return nil
}

func (me *parser) enddef() *node {
	me.eat("enddef")
	return nil
}

func (me *parser) exprDef(name string, def *node) *node {
	me.eat("id")
	fmt.Println("DEF", name, ":=", def.string(0))
	return def
}
