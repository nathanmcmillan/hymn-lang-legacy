package main

import (
	"fmt"
	"strings"
)

func (me *parser) fileExpression() {
	token := me.token
	op := token.is
	if op == "import" {
		me.eat(op)
		me.imports()
	} else if op == "immutable" {
		me.eat(op)
		me.immutables()
	} else if op == "mutable" {
		me.eat(op)
		me.mutables()
	} else if op == "id" {
		name := token.value
		if _, ok := me.hmfile.classes[name]; ok {
			me.defineClassFunction()
		} else {
			me.defineFileFunction()
		}
	} else if op == "type" {
		me.defineClass()
	} else if op == "enum" {
		me.defineEnum()
	} else if op == "def" {
		me.def()
	} else if op == "ifdef" {
		me.ifdef()
	} else if op == "elsedef" {
		me.elsedef()
	} else if op == "enddef" {
		me.enddef()
	} else if op == "comment" {
		me.eat("comment")
	} else if op == "line" || op == "eof" {
		return
	} else {
		panic(me.fail() + "unknown top level expression \"" + op + "\"")
	}
}

func (me *parser) expression() *node {
	token := me.token
	op := token.is
	if op == "mutable" {
		me.eat(op)
		n := me.forceassign(true, true)
		me.verify("line")
		return n
	}
	if op == "id" {
		name := token.value
		if _, ok := me.hmfile.functions[name]; ok {
			return me.call(me.hmfile)
		}
		n := me.eatvar(me.hmfile)
		if me.assignable(n) {
			n = me.assign(n, true, false)
		} else if n.is != "call" {
			panic(me.fail() + "expected assign or call expression for \"" + name + "\"")
		}
		me.verify("line")
		return n
	} else if op == "match" {
		return me.match()
	} else if op == "if" {
		return me.ifexpr()
	} else if op == "break" {
		return me.breaking()
	} else if op == "continue" {
		return me.continuing()
	} else if op == "for" {
		return me.forexpr()
	} else if op == "return" {
		return me.returning()
	} else if op == "goto" {
		return me.gotoLabel()
	} else if op == "label" {
		return me.label()
	} else if op == "pass" {
		return me.pass()
	} else if op == "def" {
		return me.def()
	} else if op == "ifdef" {
		return me.ifdef()
	} else if op == "elsedef" {
		return me.elsedef()
	} else if op == "enddef" {
		return me.enddef()
	} else if op == "comment" {
		return me.comment()
	} else if op == "line" || op == "eof" {
		return nil
	}
	panic(me.fail() + "unknown expression \"" + op + "\"")
}

func (me *parser) maybeIgnore(depth int) {
	for {
		if me.token.is == "line" {
			me.eat("line")
			break
		}
	}
	for me.token.is != "eof" {
		token := me.token
		if token.is != "line" {
			break
		}
		me.next()
	}
}

func (me *parser) gotoLabel() *node {
	me.eat("goto")
	n := nodeInit("goto")
	name := me.token.value
	me.eat("id")
	n.value = name
	me.verify("line")
	return n
}

func (me *parser) label() *node {
	me.eat("label")
	n := nodeInit("label")
	name := me.token.value
	me.eat("id")
	n.value = name
	me.verify("line")
	return n
}

func (me *parser) pass() *node {
	me.eat("pass")
	n := nodeInit("pass")
	me.verify("line")
	return n
}

func (me *parser) continuing() *node {
	me.eat("continue")
	n := nodeInit("continue")
	n.typed = "void"
	me.verify("line")
	return n
}

func (me *parser) breaking() *node {
	me.eat("break")
	n := nodeInit("break")
	n.typed = "void"
	me.verify("line")
	return n
}

func (me *parser) returning() *node {
	me.eat("return")
	calc := me.calc(0)
	n := nodeInit("return")
	n.typed = calc.typed
	n.push(calc)
	me.verify("line")
	return n
}

func (me *parser) forceassign(malloc, mutable bool) *node {
	v := me.eatvar(me.hmfile)
	if !me.assignable(v) {
		panic(me.fail() + "expected variable for assignment but was \"" + v.typed + "\"")
	}
	return me.assign(v, malloc, mutable)
}

func (me *parser) assign(left *node, malloc, mutable bool) *node {
	op := me.token.is
	mustBeInt := false
	mustBeNumber := false
	if op == "%=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		mustBeInt = true
	} else if op == "+=" || op == "-=" || op == "*=" || op == "/=" {
		mustBeNumber = true
	} else if op != "=" && op != ":=" {
		panic(me.fail() + "unknown assign operation \"" + op + "\"")
	}
	me.eat(op)
	right := me.calc(0)
	if mustBeInt {
		if right.typed != "int" {
			panic(me.fail() + "assign operation \"" + op + "\" requires int type")
		}
	} else if mustBeNumber {
		if !isNumber(right.typed) {
			panic(me.fail() + "assign operation \"" + op + "\" requires number type")
		}
	}
	if left.is == "variable" {
		sv := me.hmfile.getvar(left.value)
		if sv != nil {
			if !sv.mutable {
				panic(me.fail() + "variable \"" + sv.name + "\" is not mutable")
			}
		} else if mustBeInt || mustBeNumber {
			panic(me.fail() + "cannot operate \"" + op + "\" for variable \"" + left.value + "\" does not exist")
		} else {
			left.typed = right.typed
			if mutable {
				left.attributes["mutable"] = "true"
			}
			if !malloc {
				left.attributes["no-malloc"] = "true"
			}
			me.hmfile.scope.variables[left.value] = me.hmfile.varInit(right.typed, left.value, mutable, malloc)
		}
	} else if left.is == "member-variable" || left.is == "array-member" {
		if me.hmfile.typeToVarData(left.typed).notEqual(me.hmfile.typeToVarData(right.typed)) {
			if strings.HasPrefix(left.typed, right.typed) && strings.Index(left.typed, "<") != -1 {
				right.typed = left.typed
			} else {
				panic(me.fail() + "member variable type \"" + left.typed + "\" does not match expression type \"" + right.typed + "\"")
			}
		}
	} else {
		panic(me.fail() + "bad assignment \"" + left.is + "\"")
	}
	right.attributes["assign"] = left.value
	if _, useStack := right.attributes["use-stack"]; useStack {
		left.attributes["use-stack"] = "true"
	}
	n := nodeInit(op)
	if op == ":=" {
		n.typed = right.typed
	} else {
		n.typed = "void"
	}
	n.push(left)
	fmt.Println("assign set", left.string(0))
	n.push(right)
	return n
}

func (me *parser) comment() *node {
	token := me.token
	me.eat("comment")
	n := nodeInit("comment")
	n.value = token.value
	return n
}

func (me *parser) free() *node {
	me.eat("free")
	token := me.token
	me.eat("id")
	n := nodeInit("free")
	n.value = token.value
	return n
}

func (me *parser) extern() *node {
	ext := me.token
	me.eat("id")
	me.eat(".")
	extname := ext.value
	id := me.token
	if id.is != "id" {
		panic(me.fail() + "expecting id token after extern " + extname)
	}
	idname := id.value
	module := me.hmfile.program.hmfiles[extname]

	if _, ok := module.functions[idname]; ok {
		fmt.Println("extern call")
		return me.call(module)
	} else if _, ok := module.classes[idname]; ok {
		fmt.Println("extern class")
		return me.allocClass(module, nil)
	} else if _, ok := module.enums[idname]; ok {
		fmt.Println("extern enum")
		return me.allocEnum(module, nil)
	} else if module.getStatic(idname) != nil {
		fmt.Println("extern var")
		return me.eatvar(module)
	} else {
		panic(me.fail() + "external type \"" + extname + "." + idname + "\" does not exist")
	}
}

func (me *parser) block() *node {
	depth := me.token.depth
	block := nodeInit("block")
	for {
		for me.token.is == "line" {
			me.eat("line")
			if me.token.is != "line" {
				if me.token.depth < depth {
					goto blockEnd
				}
				break
			}
		}
		if me.token.is == "eof" || me.token.is == "comment" {
			goto blockEnd
		}
		expr := me.expression()
		block.push(expr)
		if expr.is == "return" {
			fn := me.hmfile.scope.fn
			if fn.typed.notEqual(me.hmfile.typeToVarData(expr.typed)) {
				panic(me.fail() + "function " + fn.name + " returns " + fn.typed.full + " but found " + expr.typed)
			}
			goto blockEnd
		}
	}
blockEnd:
	fmt.Println("> block", block.string(0))
	return block
}

func (me *parser) iswhile() bool {
	pos := me.pos
	token := me.tokens.get(pos)
	for token.is != "line" && token.is != "eof" {
		if token.is == "delim" {
			return false
		}
		pos++
		token = me.tokens.get(pos)
	}
	return true
}

func (me *parser) forexpr() *node {
	fmt.Println("> for expression")
	me.eat("for")
	n := nodeInit("for")
	n.typed = "void"
	if me.token.is == "line" {
		me.eat("line")
	} else {
		if me.iswhile() {
			fmt.Println("> regular while")
			n.push(me.calcBool())
		} else {
			fmt.Println("> multi for")
			n.push(me.forceassign(true, true))
			me.eat("delim")
			n.push(me.calcBool())
			me.eat("delim")
			n.push(me.forceassign(true, true))
		}
		me.eat("line")
	}
	n.push(me.block())
	return n
}

func (me *parser) match() *node {
	depth := me.token.depth
	fmt.Println("match depth", depth)
	me.eat("match")
	n := nodeInit("match")
	n.typed = "void"

	matching := me.calc(0)
	matchType := me.hmfile.typeToVarData(matching.typed)
	var matchVar *variable
	if matching.is == "variable" {
		matchVar = me.hmfile.getvar(matching.value)
	}

	n.push(matching)

	me.eat("line")
	for {
		if me.token.is == "id" {
			id := me.token.value
			me.eat("id")
			caseNode := nodeInit(id)
			me.eat("=>")
			n.push(caseNode)
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}

		} else if me.token.is == "_" {
			me.eat("_")
			me.eat("=>")
			n.push(nodeInit("_"))
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}

		} else if me.token.is == "some" {
			me.eat("some")
			me.eat("=>")
			n.push(nodeInit("some"))

			if matchVar != nil {
				if !matchType.maybe {
					panic("type \"" + matchVar.name + "\" is not \"maybe\"")
				}
				fmt.Println("match temp update", matchVar.name, "to", matchType.some)
				matchVar.update(me.hmfile, matchType.some)
			}

			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}

			if matchVar != nil {
				matchVar.update(me.hmfile, matchType.full)
			}

		} else if me.token.is == "none" {
			me.eat("none")
			me.eat("=>")
			n.push(nodeInit("none"))
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}

		} else if me.token.depth <= depth {
			break
		} else {
			panic(me.fail() + "unknown match expression")
		}
	}
	return n
}

func (me *parser) ifexpr() *node {
	fmt.Println("> if")
	me.eat("if")
	n := nodeInit("if")
	n.typed = "void"
	n.push(me.calcBool())
	me.eat("line")
	n.push(me.block())
	for me.token.is == "elif" {
		me.eat("elif")
		other := nodeInit("elif")
		other.push(me.calcBool())
		me.eat("line")
		other.push(me.block())
		n.push(other)
	}
	if me.token.is == "else" {
		me.eat("else")
		me.eat("line")
		n.push(me.block())
	}
	return n
}

func (me *parser) comparison(left *node, op string) *node {
	fmt.Println("> comparison " + left.string(0) + "")
	var typed string
	if op == "and" || op == "or" {
		if left.typed != "bool" {
			err := me.fail() + "left side of \"" + op + "\" must be a boolean, was \"" + left.typed + "\""
			err += "\nleft: " + left.string(0)
			panic(err)
		}
		typed = op
		me.eat(op)
	} else if op == "=" {
		typed = "equal"
		me.eat(op)
	} else if op == "!=" {
		typed = "not-equal"
		me.eat(op)
	} else if op == ">" || op == ">=" || op == "<" || op == "<=" {
		if !isNumber(left.typed) {
			err := me.fail() + "left side of comparison must be a number, was \"" + left.typed + "\""
			err += "\nleft: " + left.string(0)
			panic(err)
		}
		typed = op
		me.eat(op)
	} else {
		panic(me.fail() + "unknown token for boolean expression")
	}
	right := me.calc(0)
	if me.hmfile.typeToVarData(left.typed).notEqual(me.hmfile.typeToVarData(right.typed)) {
		panic(me.fail() + "comparison types \"" + left.typed + "\" and \"" + right.typed + "\" do not match")
	}
	n := nodeInit(typed)
	n.typed = "bool"
	n.push(left)
	n.push(right)
	fmt.Println("> compare using", typed)
	fmt.Println("> left", left.string(0))
	fmt.Println("> right", right.string(0))
	return n
}

func (me *parser) calcBool() *node {
	n := me.calc(0)
	if n.typed != "bool" {
		panic(me.fail() + "must be boolean expression")
	}
	return n
}

func (me *parser) imports() {
	me.eat("line")
	for {
		name := me.token.value
		fmt.Println("importing " + name)
		me.eat("string")
		_, ok := me.hmfile.imports[name]
		if !ok {
			me.hmfile.imports[name] = true
			path := me.hmfile.program.directory + "/" + name + ".hm"
			me.hmfile.program.compile(me.hmfile.program.out, path, me.hmfile.program.libDir)
			fmt.Println("finished compiling " + name)
			fmt.Println("=== continue " + me.hmfile.name + " parse === ")
		}
		me.eat("line")
		if me.token.is == "line" || me.token.is == "eof" || me.token.is == "comment" {
			break
		}
	}
}

func (me *parser) immutables() {
	if me.token.is == "line" {
		me.eat("line")
	}
	for {
		n := me.forceassign(false, false)
		av := n.has[0]
		fmt.Println("static immutable", n.string(0))
		if n.is != "=" || av.is != "variable" {
			panic(me.fail() + "invalid static variable")
		}
		me.hmfile.statics = append(me.hmfile.statics, n)
		me.hmfile.staticScope[av.value] = me.hmfile.scope.variables[av.value]
		me.eat("line")
		if me.token.is == "line" || me.token.is == "eof" || me.token.is == "comment" {
			break
		}
	}
}

func (me *parser) mutables() {
	me.eat("line")
	for {
		n := me.forceassign(false, true)
		av := n.has[0]
		fmt.Println("static mutable", n.string(0))
		if n.is != "=" || av.is != "variable" {
			panic(me.fail() + "invalid static variable")
		}
		me.hmfile.statics = append(me.hmfile.statics, n)
		me.hmfile.staticScope[av.value] = me.hmfile.scope.variables[av.value]
		me.eat("line")
		if me.token.is == "line" || me.token.is == "eof" || me.token.is == "comment" {
			break
		}
	}
}
