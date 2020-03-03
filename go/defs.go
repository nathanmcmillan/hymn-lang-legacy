package main

import "fmt"

func (me *parser) macro() *parseError {
	me.eat("macro")
	name := me.token.value
	me.eat("id")
	var value *node
	if me.token.is != "line" {
		var er *parseError
		value, er = me.calc(0, nil)
		if er != nil {
			return er
		}
		fmt.Println("NEW DEF IS", name, ":=", value.string(me.hmfile, 0))
	} else {
		fmt.Println("NEW DEF IS", name)
	}
	me.eat("line")
	me.hmfile.defs[name] = value
	me.hmfile.namespace[name] = "def"
	me.hmfile.types[name] = "def"
	return nil
}

func (me *parser) ifdef() *parseError {
	me.eat("ifdef")
	name := me.token.value
	me.eat("id")
	if _, ok := me.hmfile.defs[name]; ok {
		for {
			if me.token.is == "elsedef" || me.token.is == "enddef" {
				break
			}
			if me.token.is == "eof" {
				return err(me, ECodeUnexpectedToken, "ifdef "+name+" missing enddef")
			}
		}
	} else {
		for {
			if me.token.is == "elsedef" || me.token.is == "enddef" {
				break
			}
			if me.token.is == "eof" {
				return err(me, ECodeUnexpectedToken, "ifdef "+name+" missing enddef")
			}
		}
	}
	return nil
}

func (me *parser) elsedef() *parseError {
	me.eat("elsedef")
	return nil
}

func (me *parser) enddef() *parseError {
	me.eat("enddef")
	return nil
}

func (me *parser) exprDef(name string, def *node) (*node, *parseError) {
	me.eat("id")
	fmt.Println("DEF", name, ":=", def.string(me.hmfile, 0))
	return def, nil
}
