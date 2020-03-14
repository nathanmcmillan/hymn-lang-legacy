package main

import (
	"fmt"
	"os"
	"strings"
)

func (me *parser) statement() *parseError {
	token := me.token
	op := token.is
	if op == "import" {
		return me.importing()
	} else if op == "const" {
		return me.immutable()
	} else if op == "mutable" {
		return me.mutable()
	} else if op == "def" {
		return me.defineNewFunction()
	} else if op == "class" {
		return me.defineClass()
	} else if op == "interface" {
		return me.defineInterface()
	} else if op == "enum" {
		return me.defineEnum()
	} else if op == "macro" {
		return me.macro()
	} else if op == "ifdef" {
		return me.ifdef()
	} else if op == "elsedef" {
		return me.elsedef()
	} else if op == "enddef" {
		return me.enddef()
	} else if op == "comment" {
		me.topComment()
		return nil
	} else if op == "line" || op == "eof" {
		return nil
	}
	expr, er := me.expression()
	if er != nil {
		return er
	}
	me.hmfile.top = append(me.hmfile.top, expr)
	return nil
}

func (me *parser) expression() (*node, *parseError) {
	token := me.token
	op := token.is
	if op == "mutable" {
		return me.parseMutable()
	} else if op == "id" {
		return me.parseIdent()
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
	} else if op == "macro" {
		return nil, me.macro()
	} else if op == "ifdef" {
		return nil, me.ifdef()
	} else if op == "elsedef" {
		return nil, me.elsedef()
	} else if op == "enddef" {
		return nil, me.enddef()
	} else if op == "comment" {
		return me.comment()
	} else if op == "line" || op == "eof" {
		return nil, nil
	}
	return nil, err(me, ECodeUnexpectedToken, "Unknown token '"+op+"'")
}

func (me *parser) parseMutable() (*node, *parseError) {
	if er := me.eat("mutable"); er != nil {
		return nil, er
	}
	ev, er := me.eatvar(me.hmfile)
	if er != nil {
		return nil, er
	}
	n, er := me.forceassign(ev, true, true)
	if er != nil {
		return nil, er
	}
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) parseIdent() (*node, *parseError) {

	name := me.token.value
	module := me.hmfile

	if _, ok := module.imports[name]; ok && me.peek().is == "." {
		return me.extern()
	}

	if _, ok := module.getType(name); ok {
		if _, ok := module.getFunction(name); ok {
			return me.parseFn(module)
		}
		return nil, err(me, ECodeExpectingExpression, "Type '"+name+"' must be assigned.")
	}

	n, er := me.eatvar(me.hmfile)
	if er != nil {
		return nil, er
	}

	if me.assignable(n) {
		n, er = me.assign(n, true, false)
		if er != nil {
			return nil, er
		}
	} else if n.is != "call" {
		return nil, err(me, ECodeExpectingExpression, "Expected assignment or call expression for '"+name+"'")
	}

	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) maybeIgnore(depth int) *parseError {
	for {
		if me.token.is == "line" {
			if er := me.eat("line"); er != nil {
				return er
			}
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
	return nil
}

func (me *parser) gotoLabel() (*node, *parseError) {
	if er := me.eat("goto"); er != nil {
		return nil, er
	}
	n := nodeInit("goto")
	name := me.token.value
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	n.value = name
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) label() (*node, *parseError) {
	if er := me.eat("label"); er != nil {
		return nil, er
	}
	n := nodeInit("label")
	name := me.token.value
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	n.value = name
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) pass() (*node, *parseError) {
	if er := me.eat("pass"); er != nil {
		return nil, er
	}
	n := nodeInit("pass")
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) continuing() (*node, *parseError) {
	if er := me.eat("continue"); er != nil {
		return nil, er
	}
	n := nodeInit("continue")
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) breaking() (*node, *parseError) {
	if er := me.eat("break"); er != nil {
		return nil, er
	}
	n := nodeInit("break")
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) parseReturn() (*node, *parseError) {
	if er := me.eat("return"); er != nil {
		return nil, er
	}
	n := nodeInit("return")
	if me.token.is != "line" {
		fn := me.hmfile.scope.fn
		calc, er := me.calc(0, fn.returns)
		if er != nil {
			return nil, er
		}
		n.copyDataOfNode(calc)
		n.push(calc)
		ret := calc.data()
		if ret.isNone() {
			if !fn.returns.isSome() {
				return nil, err(me, ECodeReturnTypeMismatch, "return type was \""+ret.print()+"\" but function is \""+fn.returns.print()+"\"")
			} else if ret.getmember() != nil {
				if calc.is == "none" {
					return nil, err(me, ECodeRedundantNoneDefinition, "unnecessary none definition for return "+calc.string(me.hmfile, 0))
				}
			}
		} else if fn.returns.notEquals(ret) {
			return nil, err(me, ECodeReturnTypeMismatch, "Function "+fn.canonical(me.hmfile)+" returns "+fn.returns.error()+" but found "+ret.error())
		}
	}
	if er := me.verify("line"); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) defineNewFunction() *parseError {
	if er := me.eat("def"); er != nil {
		return er
	}
	name := me.token.value
	if _, ok := me.hmfile.classes[name]; ok {
		return me.defineClassFunction()
	}
	return me.defineStaticFunction()
}

func (me *parser) comment() (*node, *parseError) {
	token := me.token
	if er := me.eat("comment"); er != nil {
		return nil, er
	}
	n := nodeInit("comment")
	n.value = token.value
	return n, nil
}

func (me *parser) topComment() *parseError {
	token := me.token
	if er := me.eat("comment"); er != nil {
		return er
	}
	me.hmfile.comments = append(me.hmfile.comments, token.value)
	return nil
}

func (me *parser) free() (*node, *parseError) {
	if er := me.eat("free"); er != nil {
		return nil, er
	}
	token := me.token
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	n := nodeInit("free")
	n.value = token.value
	return n, nil
}

func (me *parser) extern() (*node, *parseError) {
	token1 := me.token
	if er := me.verify("id"); er != nil {
		return nil, er
	}
	token2 := me.peek()
	token3 := me.doublePeek()

	name := token1.value
	en, possiblyAnEnum := me.hmfile.enums[name]

	if token2.is != "." {
		var e string
		if possiblyAnEnum {
			e = fmt.Sprintf("I expected `.` to follow after the module or enum `%s`", name)
		} else {
			e = fmt.Sprintf("I expected `.` to follow after the module `%s`", name)
		}
		return nil, err(me, ECodeUnexpectedToken, e)
	}

	if token3.is != "id" {
		var e string
		if possiblyAnEnum {
			e = fmt.Sprintf("I expected an identifier to follow after the module or enum `%s`", name)
		} else {
			e = fmt.Sprintf("I expected an identifier to follow after the module `%s`", name)
		}
		return nil, err(me, ECodeUnexpectedToken, e)
	}

	id := token3.value

	if possiblyAnEnum {
		if un := en.getType(id); un != nil {
			return me.allocEnum(me.hmfile, nil)
		}
	}

	me.next()
	me.next()
	module := me.hmfile.imports[name]

	if _, ok := module.functions[id]; ok {
		return me.parseFn(module)
	} else if _, ok := module.classes[id]; ok {
		return me.allocClass(module, nil)
	} else if _, ok := module.enums[id]; ok {
		return me.allocEnum(module, nil)
	} else if module.getStatic(id) != nil {
		return me.eatvar(module)
	} else {
		e := fmt.Sprintf("I could not find the external type `%s.%s`", name, id)
		return nil, err(me, ECodeUnknownType, e)
	}
}

func (me *parser) block() (*node, *parseError) {
	depth := me.token.depth
	block := nodeInit("block")
	for {
		for me.token.is == "line" {
			if er := me.eat("line"); er != nil {
				return nil, er
			}
		}
		if me.token.depth < depth || me.token.is == "eof" || me.token.is == "comment" {
			goto blockEnd
		}
		e, er := me.expression()
		if er != nil {
			return nil, er
		}
		block.push(e)
	}
blockEnd:
	return block, nil
}

func (me *parser) calcBool() (*node, *parseError) {
	n, er := me.calc(0, nil)
	if er != nil {
		return nil, er
	}
	if !n.data().isBoolean() {
		return nil, err(me, ECodeBooleanRequired, "must be boolean expression")
	}
	return n, nil
}

func (me *parser) global(mutable bool) *parseError {
	module := me.hmfile
	v, er := me.eatvar(me.hmfile)
	if er != nil {
		return er
	}
	name := v.idata.name
	existing := module.getvar(name)
	if existing != nil {
		return err(me, ECodeNameConflict, "Variable \""+name+"\" already defined.")
	}
	v.idata.setGlobal(true)
	n, er := me.forceassign(v, true, mutable)
	if er != nil {
		return er
	}
	module.statics = append(module.statics, n)
	module.staticScope[name] = &variableNode{n, me.hmfile.scope.variables[name]}
	if er := me.eat("line"); er != nil {
		return er
	}
	return nil
}

func (me *parser) immutable() *parseError {
	if er := me.eat("const"); er != nil {
		return er
	}
	return me.global(false)
}

func (me *parser) mutable() *parseError {
	if er := me.eat("mutable"); er != nil {
		return er
	}
	return me.global(true)
}

func variableSubstitution(value string, variables map[string]string) string {
	expanded := ""
	index := 0
	remainder := value
	for true {
		sign := strings.Index(remainder, "$")
		if sign == -1 {
			break
		}
		left := strings.Index(remainder, "{")
		if left != sign+1 {
			break
		}
		right := strings.Index(remainder, "}")
		if right <= left+1 {
			break
		}
		expanded += remainder[0:sign]
		variable := remainder[left+1 : right]
		var env string
		if v, ok := variables[variable]; ok {
			env = v
		} else {
			env = os.Getenv(variable)
		}
		if env != "" {
			expanded += env
		}
		index = right + 1
		remainder = remainder[index:]
	}
	if index < len(value)-1 {
		expanded += remainder
	}
	return expanded
}
