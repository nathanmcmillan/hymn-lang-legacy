package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (me *parser) fileExpression() *parseError {
	token := me.token
	op := token.is
	if op == "import" {
		return me.importing()
	} else if op == "const" {
		return me.immutable()
	} else if op == "mutable" {
		return me.mutable()
	} else if op == "def" || op == "id" {
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
	} else {
		return err(me, "unknown top level expression \""+op+"\"")
	}
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
		return me.comment(), nil
	} else if op == "line" || op == "eof" {
		return nil, nil
	}
	return nil, err(me, "Unknown token '"+op+"'")
}

func (me *parser) parseMutable() (*node, *parseError) {
	me.eat("mutable")
	ev, er := me.eatvar(me.hmfile)
	if er != nil {
		return nil, er
	}
	n, er := me.forceassign(ev, true, true)
	if er != nil {
		return nil, er
	}
	me.verify("line")
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
		return nil, err(me, "Type '"+name+"' must be assigned.")
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
		return nil, err(me, "Expected assignment or call expression for '"+name+"'")
	}

	me.verify("line")
	return n, nil
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

func (me *parser) gotoLabel() (*node, *parseError) {
	me.eat("goto")
	n := nodeInit("goto")
	name := me.token.value
	me.eat("id")
	n.value = name
	me.verify("line")
	return n, nil
}

func (me *parser) label() (*node, *parseError) {
	me.eat("label")
	n := nodeInit("label")
	name := me.token.value
	me.eat("id")
	n.value = name
	me.verify("line")
	return n, nil
}

func (me *parser) pass() (*node, *parseError) {
	me.eat("pass")
	n := nodeInit("pass")
	me.verify("line")
	return n, nil
}

func (me *parser) continuing() (*node, *parseError) {
	me.eat("continue")
	n := nodeInit("continue")
	me.verify("line")
	return n, nil
}

func (me *parser) breaking() (*node, *parseError) {
	me.eat("break")
	n := nodeInit("break")
	me.verify("line")
	return n, nil
}

func (me *parser) parseReturn() (*node, *parseError) {
	me.eat("return")
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
				return nil, err(me, "return type was \""+ret.print()+"\" but function is \""+fn.returns.print()+"\"")
			} else if ret.getmember() != nil {
				if calc.is == "none" {
					return nil, err(me, "unnecessary none definition for return "+calc.string(me.hmfile, 0))
				}
			}
		} else if fn.returns.notEquals(ret) {
			return nil, err(me, "Function "+fn.canonical(me.hmfile)+" returns "+fn.returns.error()+" but found "+ret.error())
		}
	}
	me.verify("line")
	return n, nil
}

func (me *parser) defineNewFunction() *parseError {
	if me.token.is == "def" {
		me.eat("def")
	}
	name := me.token.value
	if _, ok := me.hmfile.classes[name]; ok {
		return me.defineClassFunction()
	}
	return me.defineStaticFunction()
}

func (me *parser) comment() *node {
	token := me.token
	me.eat("comment")
	n := nodeInit("comment")
	n.value = token.value
	return n
}

func (me *parser) topComment() {
	token := me.token
	me.eat("comment")
	me.hmfile.comments = append(me.hmfile.comments, token.value)
}

func (me *parser) free() *node {
	me.eat("free")
	token := me.token
	me.eat("id")
	n := nodeInit("free")
	n.value = token.value
	return n
}

func (me *parser) extern() (*node, *parseError) {
	ext := me.token
	me.eat("id")
	if me.token.is != "." {
		return nil, err(me, "expecting \".\" after module name")
	}
	me.eat(".")
	extname := ext.value
	id := me.token
	if id.is != "id" {
		return nil, err(me, "expecting id token after extern "+extname)
	}
	idname := id.value
	module := me.hmfile.imports[extname]

	if _, ok := module.functions[idname]; ok {
		return me.parseFn(module)
	} else if _, ok := module.classes[idname]; ok {
		return me.allocClass(module, nil)
	} else if _, ok := module.enums[idname]; ok {
		return me.allocEnum(module, nil)
	} else if module.getStatic(idname) != nil {
		return me.eatvar(module)
	} else {
		return nil, err(me, "external type \""+extname+"."+idname+"\" does not exist")
	}
}

func (me *parser) block() (*node, *parseError) {
	depth := me.token.depth
	block := nodeInit("block")
	for {
		for me.token.is == "line" {
			me.eat("line")
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
		return nil, err(me, "must be boolean expression")
	}
	return n, nil
}

func (me *parser) importing() *parseError {
	me.eat("import")
	value := me.token.value
	me.eat(TokenStringLiteral)
	value = variableSubstitution(value, me.hmfile.program.shellvar)
	absolute, er := filepath.Abs(value)
	if er != nil {
		return err(me, "Failed to parse import \""+value+"\". "+er.Error())
	}
	alias := filepath.Base(absolute)
	if me.token.is == "as" {
		me.eat("as")
		alias = me.token.value
		me.eat("id")
	}
	statics := make([]string, 0)
	if me.token.is == "(" {
		me.eat("(")
		if me.token.is == "line" {
			me.eat("line")
		}
		for me.token.is != ")" {
			value := me.token.value
			me.eat("id")
			statics = append(statics, value)
			if me.token.is == "line" {
				me.eat("line")
			} else if me.token.is == "," {
				me.eat(",")
			}
		}
		me.eat(")")
	}

	module := me.hmfile

	path, er := filepath.Abs(filepath.Join(module.program.directory, value+".hm"))
	if er != nil {
		return err(me, "Failed to parse import \""+value+"\". "+er.Error())
	}

	var importing *hmfile
	found, ok := module.program.hmfiles[path]
	if ok {
		if _, ok := module.importPaths[path]; ok {
			return err(me, "Module \""+path+"\" was already imported.")
		}
		importing = found
	} else {
		out, fer := filepath.Abs(filepath.Join(module.program.out, value))
		if fer != nil {
			return err(me, "Failed to parse import \""+value+"\". "+er.Error())
		}

		var er *parseError
		importing, er = module.program.parse(out, path, module.program.libs)
		if er != nil {
			return er
		}

		if debug {
			fmt.Println("=== parse: " + module.name + " ===")
		}
	}

	module.imports[alias] = importing
	module.importPaths[path] = importing
	module.importOrder = append(module.importOrder, alias)
	importing.crossref[module] = alias

	for _, s := range statics {
		if cl, ok := importing.classes[s]; ok {
			if _, ok := module.types[cl.name]; ok {
				return err(me, "Cannot import class \""+cl.name+"\". It is already defined.")
			}
			module.classes[cl.name] = cl
			module.namespace[cl.name] = "class"
			module.types[cl.name] = "class"

			module.classes[cl.uid()] = cl
			module.namespace[cl.uid()] = "class"
			module.types[cl.uid()] = "class"

		} else if in, ok := importing.interfaces[s]; ok {
			if _, ok := module.types[in.name]; ok {
				return err(me, "Cannot import interface \""+in.name+"\". It is already defined.")
			}
			module.interfaces[in.name] = in
			module.namespace[in.name] = "interface"
			module.types[in.name] = "interface"

			module.interfaces[in.uid()] = in
			module.namespace[in.uid()] = "interface"
			module.types[in.uid()] = "interface"

		} else if en, ok := importing.enums[s]; ok {
			if _, ok := module.types[en.name]; ok {
				return err(me, "Cannot import enum \""+en.name+"\". It is already defined.")
			}
			module.enums[en.name] = en
			module.namespace[en.name] = "enum"
			module.types[en.name] = "enum"

			module.enums[en.uid()] = en
			module.namespace[en.uid()] = "enum"
			module.types[en.uid()] = "enum"

		} else if fn, ok := importing.functions[s]; ok {
			name := fn.getname()
			if _, ok := module.types[name]; ok {
				return err(me, "Cannot import function \""+name+"\". It is already defined.")
			}
			module.functions[name] = fn
			module.namespace[name] = "function"
			module.types[name] = "function"

		} else if st, ok := importing.staticScope[s]; ok {
			if _, ok := module.types[st.v.name]; ok {
				return err(me, "Cannot import variable \""+st.v.name+"\". It is already defined.")
			}
			module.staticScope[st.v.name] = st
			module.scope.variables[st.v.name] = st.v
		}
	}

	me.eat("line")

	return nil
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
		return err(me, "Variable \""+name+"\" already defined.")
	}
	v.idata.setGlobal(true)
	n, er := me.forceassign(v, true, mutable)
	if er != nil {
		return er
	}
	module.statics = append(module.statics, n)
	module.staticScope[name] = &variableNode{n, me.hmfile.scope.variables[name]}
	me.eat("line")
	return nil
}

func (me *parser) immutable() *parseError {
	me.eat("const")
	return me.global(false)
}

func (me *parser) mutable() *parseError {
	me.eat("mutable")
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
