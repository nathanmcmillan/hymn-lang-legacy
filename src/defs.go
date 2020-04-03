package main

import "fmt"

func (me *parser) macro() *parseError {
	if er := me.eat("macro"); er != nil {
		return er
	}
	name := me.token.value
	if er := me.eat("id"); er != nil {
		return er
	}
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
	if er := me.newLine(); er != nil {
		return er
	}
	me.hmfile.defs[name] = value
	me.hmfile.namespace[name] = "def"
	me.hmfile.types[name] = "def"
	return nil
}

func (me *parser) ifdef() *parseError {
	if er := me.eat("ifdef"); er != nil {
		return er
	}
	name := me.token.value
	if er := me.eat("id"); er != nil {
		return er
	}
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
	if er := me.eat("elsedef"); er != nil {
		return er
	}
	return nil
}

func (me *parser) enddef() *parseError {
	if er := me.eat("enddef"); er != nil {
		return er
	}
	return nil
}

func (me *parser) exprDef(name string, def *node) (*node, *parseError) {
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	fmt.Println("DEF", name, ":=", def.string(me.hmfile, 0))
	return def, nil
}
