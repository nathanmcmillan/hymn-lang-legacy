package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (me *parser) fileExpression() {
	token := me.token
	op := token.is
	if op == "import" {
		me.importing()
	} else if op == "const" {
		me.immutable()
	} else if op == "mutable" {
		me.mutable()
	} else if op == "function" || op == "id" {
		me.defineNewFunction()
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
		me.topComment()
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
		n := me.forceassign(me.eatvar(me.hmfile), true, true)
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
		fn := me.hmfile.scope.fn
		calc := me.calc(0, fn.returns)
		n.copyDataOfNode(calc)
		n.push(calc)
		ret := calc.data()
		if ret.isNone() {
			if !fn.returns.isSome() {
				panic(me.fail() + "return type was \"" + ret.print() + "\" but function is \"" + fn.returns.print() + "\"")
			} else if ret.getmember() != nil {
				if calc.is == "none" {
					panic(me.fail() + "unnecessary none definition for return " + calc.string(me.hmfile, 0))
				}
			}
		} else if fn.returns.notEquals(ret) {
			panic(me.fail() + "Function " + fn.canonical(me.hmfile) + " returns " + fn.returns.error() + " but found " + ret.error())
		}
	}
	me.verify("line")
	return n
}

func (me *parser) defineNewFunction() {
	if me.token.is == "function" {
		me.eat("function")
	}
	name := me.token.value
	if _, ok := me.hmfile.classes[name]; ok {
		me.defineClassFunction()
	} else {
		me.defineStaticFunction()
	}
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

func (me *parser) extern() *node {
	ext := me.token
	me.eat("id")
	if me.token.is != "." {
		panic(me.fail() + "expecting \".\" after module name")
	}
	me.eat(".")
	extname := ext.value
	id := me.token
	if id.is != "id" {
		panic(me.fail() + "expecting id token after extern " + extname)
	}
	idname := id.value
	module := me.hmfile.imports[extname]

	if _, ok := module.functions[idname]; ok {
		return me.parseFn(module)
	} else if _, ok := module.classes[idname]; ok {
		return me.allocClass(module, nil)
	} else if _, ok := module.enums[idname]; ok {
		return me.allocEnum(module)
	} else if module.getStatic(idname) != nil {
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
	n := me.calc(0, nil)
	if !n.data().isBoolean() {
		panic(me.fail() + "must be boolean expression")
	}
	return n
}

func (me *parser) importing() {
	me.eat("import")
	value := me.token.value
	me.eat(TokenStringLiteral)
	value = variableSubstitution(value, me.hmfile.program.shellvar)
	absolute, err := filepath.Abs(value)
	if err != nil {
		panic(me.fail() + "Failed to parse import \"" + value + "\". " + err.Error())
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

	path, err := filepath.Abs(filepath.Join(module.program.directory, value+".hm"))
	if err != nil {
		panic(me.fail() + "Failed to parse import \"" + value + "\". " + err.Error())
	}

	var importing *hmfile
	found, ok := module.program.hmfiles[path]
	if ok {
		if _, ok := module.importPaths[path]; ok {
			panic(me.fail() + "Module \"" + path + "\" was already imported.")
		}
		importing = found
	} else {
		out, err := filepath.Abs(filepath.Join(module.program.out, value))
		if err != nil {
			panic(me.fail() + "Failed to parse import \"" + value + "\". " + err.Error())
		}

		importing = module.program.parse(out, path, module.program.libs)
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
				panic(me.fail() + "Cannot import class \"" + cl.name + "\". It is already defined.")
			}
			module.classes[cl.name] = cl
			module.namespace[cl.name] = "class"
			module.types[cl.name] = "class"

			module.classes[cl.uid()] = cl
			module.namespace[cl.uid()] = "class"
			module.types[cl.uid()] = "class"

		} else if en, ok := importing.enums[s]; ok {
			if _, ok := module.types[en.name]; ok {
				panic(me.fail() + "Cannot import enum \"" + cl.name + "\". It is already defined.")
			}
			module.enums[en.name] = en
			module.namespace[en.name] = "enum"
			module.types[en.name] = "enum"

			module.enums[en.uid()] = en
			module.namespace[en.uid()] = "enum"
			module.types[en.uid()] = "enum"

		} else if fn, ok := importing.functions[s]; ok {
			if _, ok := module.types[fn.name]; ok {
				panic(me.fail() + "Cannot import function \"" + cl.name + "\". It is already defined.")
			}
			module.functions[fn.name] = fn
			module.namespace[fn.name] = "function"
			module.types[fn.name] = "function"

		} else if st, ok := importing.staticScope[s]; ok {
			if _, ok := module.types[st.v.name]; ok {
				panic(me.fail() + "Cannot import variable \"" + st.v.name + "\". It is already defined.")
			}
			module.staticScope[st.v.name] = st
			module.scope.variables[st.v.name] = st.v
		}
	}

	me.eat("line")
}

func (me *parser) global(mutable bool) {
	module := me.hmfile
	v := me.eatvar(me.hmfile)
	name := v.idata.name
	existing := module.getvar(name)
	if existing != nil {
		panic(me.fail() + "Variable \"" + name + "\" already defined.")
	}
	v.idata.setGlobal(true)
	n := me.forceassign(v, true, mutable)
	module.statics = append(module.statics, n)
	module.staticScope[name] = &variableNode{n, me.hmfile.scope.variables[name]}
	me.eat("line")
}

func (me *parser) immutable() {
	me.eat("const")
	me.global(false)
}

func (me *parser) mutable() {
	me.eat("mutable")
	me.global(true)
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
