package main

import (
	"fmt"
	"strings"
)

func (me *parser) fileExpression() {
	token := me.token
	op := token.is
	if op == "import" {
		me.importing()
	} else if op == "immutable" {
		me.eat(op)
		me.immutable()
	} else if op == "mutable" {
		me.eat(op)
		me.mutable()
	} else if op == "id" {
		name := token.value
		if _, ok := me.hmfile.classes[name]; ok {
			me.defineClassFunction()
		} else {
			me.defineStaticFunction()
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
		if _, ok := me.hmfile.getFunction(name); ok {
			return me.parseFn(me.hmfile)
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
		return me.parseMatch()
	} else if op == "if" {
		return me.ifexpr()
	} else if op == "break" {
		return me.breaking()
	} else if op == "continue" {
		return me.continuing()
	} else if op == "for" {
		return me.forloop()
	} else if op == "while" {
		return me.whileloop()
	} else if op == "iterate" {
		return me.iterloop()
	} else if op == "return" {
		return me.parseReturn()
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
	me.verify("line")
	return n
}

func (me *parser) breaking() *node {
	me.eat("break")
	n := nodeInit("break")
	me.verify("line")
	return n
}

func (me *parser) parseReturn() *node {
	me.eat("return")
	n := nodeInit("return")
	if me.token.is != "line" {
		calc := me.calc(0)
		n.copyDataOfNode(calc)
		n.push(calc)
		fn := me.hmfile.scope.fn
		ret := calc.data()
		if ret.none {
			if !fn.returns.maybe {
				panic(me.fail() + "return type was \"" + ret.full + "\" but function is \"" + fn.returns.full + "\"")
			} else if ret.memberType.full != "" {
				if calc.is == "none" {
					panic(me.fail() + "unnecessary none definition for return " + calc.string(0))
				}
			}
		} else if fn.returns.notEqual(ret) {
			panic(me.fail() + "function \"" + fn.canonical() + "\" returns \"" + fn.returns.full + "\" but found \"" + calc.getType() + "\"")
		}
	}
	me.verify("line")
	return n
}

func (me *parser) forceassign(malloc, mutable bool) *node {
	v := me.eatvar(me.hmfile)
	if !me.assignable(v) {
		panic(me.fail() + "expected variable for assignment but was \"" + v.getType() + "\"")
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
	me.hmfile.assignmentStack = append(me.hmfile.assignmentStack, left)
	right := me.calc(0)
	me.hmfile.assignmentStack = me.hmfile.assignmentStack[0 : len(me.hmfile.assignmentStack)-1]
	if mustBeInt {
		if right.getType() != TokenInt {
			panic(me.fail() + "assign operation \"" + op + "\" requires int type")
		}
	} else if mustBeNumber {
		if !isNumber(right.getType()) {
			panic(me.fail() + "assign operation \"" + op + "\" requires number type")
		}
	}
	if left.is == "variable" {
		sv := me.hmfile.getvar(left.idata.name)
		if sv != nil {
			if !sv.mutable {
				panic(me.fail() + "variable \"" + sv.name + "\" is not mutable")
			}
			if right.data().full != "?" && left.data().notEqual(right.data()) {
				if strings.HasPrefix(left.getType(), right.getType()) && strings.Index(left.getType(), "<") != -1 {
					right.copyDataOfNode(left)
				} else {
					fmt.Println(left.string(0), "->", left.data().dtype.standard())
					fmt.Println(right.string(0), "->", right.data().dtype.standard())
					panic(me.fail() + "variable type \"" + left.getType() + "\" does not match expression type \"" + right.getType() + "\"")
				}
			}
		} else if mustBeInt || mustBeNumber {
			panic(me.fail() + "cannot operate \"" + op + "\" for variable \"" + left.idata.name + "\" does not exist")
		} else {
			left.copyDataOfNode(right)
			if mutable {
				left.attributes["mutable"] = "true"
			}
			if !malloc {
				left.attributes["global"] = "true"
				right.data().isptr = false
			}
			me.hmfile.scope.variables[left.idata.name] = me.hmfile.varInitFromData(right.data(), left.idata.name, mutable)
		}
	} else if left.is == "member-variable" || left.is == "array-member" {
		if right.data().full != "?" && left.data().notEqual(right.data()) {
			if strings.HasPrefix(left.getType(), right.getType()) && strings.Index(left.getType(), "<") != -1 {
				right.copyDataOfNode(left)
			} else {
				panic(me.fail() + "member variable type \"" + left.getType() + "\" does not match expression type \"" + right.getType() + "\"")
			}
		}
	} else {
		panic(me.fail() + "bad assignment \"" + left.is + "\"")
	}
	if left.idata != nil && left.is == "variable" {
		right.attributes["assign"] = left.idata.name
	}
	if _, useStack := right.attributes["stack"]; useStack {
		left.attributes["stack"] = "true"
	}
	n := nodeInit(op)
	if op == ":=" {
		n.copyDataOfNode(right)
	}
	n.push(left)
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
		return me.parseFn(module)
	} else if _, ok := module.classes[idname]; ok {
		fmt.Println("extern class", extname, idname)
		return me.allocClass(module, nil)
	} else if _, ok := module.enums[idname]; ok {
		fmt.Println("extern enum")
		return me.allocEnum(module)
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
		}
		if me.token.depth < depth || me.token.is == "eof" || me.token.is == "comment" {
			goto blockEnd
		}
		block.push(me.expression())
	}
blockEnd:
	return block
}

func (me *parser) calcBool() *node {
	n := me.calc(0)
	if n.getType() != "bool" {
		panic(me.fail() + "must be boolean expression")
	}
	return n
}

func (me *parser) importing() {
	me.eat("import")
	name := me.token.value
	fmt.Println("importing " + name)
	me.eat(TokenStringLiteral)
	module := me.hmfile
	_, ok := module.imports[name]
	if !ok {
		module.imports[name] = true
		module.importOrder = append(module.importOrder, name)
		out := module.program.out + "/" + name
		path := module.program.directory + "/" + name + ".hm"
		module.program.compile(out, path, module.program.libDir)
		fmt.Println("=== continuing " + module.name + " === ")
	}
	if me.token.is == "id" {
		fmt.Println("include specific type/enum/function/variable")
	} else if me.token.is == "*" {
		fmt.Println("include all of package")
	}
	me.eat("line")
}

func (me *parser) immutable() {
	n := me.forceassign(false, false)
	av := n.has[0]
	if n.is != "=" || av.is != "variable" {
		panic(me.fail() + "invalid static variable")
	}
	me.hmfile.statics = append(me.hmfile.statics, n)
	me.hmfile.staticScope[av.idata.name] = me.hmfile.scope.variables[av.idata.name]
	me.eat("line")
}

func (me *parser) mutable() {
	n := me.forceassign(false, true)
	av := n.has[0]
	fmt.Println("static mutable", n.string(0))
	if n.is != "=" || av.is != "variable" {
		panic(me.fail() + "invalid static variable")
	}
	me.hmfile.statics = append(me.hmfile.statics, n)
	me.hmfile.staticScope[av.idata.name] = me.hmfile.scope.variables[av.idata.name]
	me.eat("line")
}
