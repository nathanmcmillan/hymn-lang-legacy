package main

import (
	"fmt"
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
			me.classfunction()
		} else {
			me.filefunction()
		}
	} else if op == "import" {
		return
	} else if op == "class" {
		me.defineClass()
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
	} else if op == "free" {
		me.free()
		return nil
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

func (me *parser) classfunction() {
	module := me.hmfile
	className := me.token.value
	me.eat("id")
	funcName := me.token.value
	globalFuncName := me.nameOfClassFunc(className, funcName)
	me.eat("id")
	class := module.classes[className]
	if _, ok := module.functions[globalFuncName]; ok {
		panic(me.fail() + "class \"" + className + "\" with function \"" + funcName + "\" is already defined")
	}
	if _, ok := class.variables[funcName]; ok {
		panic(me.fail() + "class \"" + className + "\" with variable \"" + funcName + "\" is already defined")
	}
	fn := me.function(funcName, class)
	module.functionOrder = append(module.functionOrder, globalFuncName)
	module.functions[globalFuncName] = fn
}

func (me *parser) filefunction() {
	program := me.hmfile
	token := me.token
	name := token.value
	if _, ok := program.functions[name]; ok {
		panic(me.fail() + "function \"" + name + "\" is already defined")
	}
	me.eat("id")
	fn := me.function(name, nil)
	program.functionOrder = append(program.functionOrder, name)
	program.functions[name] = fn
}

func (me *parser) function(name string, self *class) *function {
	fn := funcInit()
	fn.name = name
	fn.typed = "void"
	if self != nil {
		ref := me.hmfile.varInit(self.name, "self", false, true)
		fn.args = append(fn.args, ref)
	}
	if me.token.is == "(" {
		me.eat("(")
		if me.token.is != ")" {
			for {
				if me.token.is == "id" {
					argname := me.token.value
					me.eat("id")
					typed := me.typedecl()
					fn.args = append(fn.args, me.hmfile.varInit(typed, argname, false, true))
					if me.token.is == ")" {
						break
					} else if me.token.is == "delim" {
						me.eat("delim")
						continue
					}
				}
				panic(me.fail() + "unexpected token in function definition")
			}
		}
		me.eat(")")
		if me.token.is != "line" {
			fn.typed = me.typedecl()
		}
		me.eat("line")
	} else {
		me.eat("line")
	}
	me.hmfile.pushScope()
	me.hmfile.scope.fn = fn
	for _, arg := range fn.args {
		me.hmfile.scope.variables[arg.name] = arg
	}
	for {
		token := me.token
		done := false
		for token.is == "line" {
			me.eat("line")
			token = me.token
			if token.is != "line" {
				if token.depth != 1 {
					done = true
				}
				break
			}
		}
		if done {
			break
		}
		if token.is == "eof" {
			break
		}
		expr := me.expression()
		fn.expressions = append(fn.expressions, expr)
		if expr.is == "return" {
			if fn.typed != expr.typed {
				panic(me.fail() + "function " + name + " returns " + fn.typed + " but found " + expr.typed)
			}
			break
		}
	}
	me.hmfile.popScope()
	return fn
}

func (me *parser) returning() *node {
	me.eat("return")
	calc := me.calc()
	n := nodeInit("return")
	n.typed = calc.typed
	n.push(calc)
	me.verify("line")
	return n
}

func (me *parser) pushparam(call *node, arg *variable) {
	param := me.factor()
	if param.typed != arg.typed && arg.typed != "?" {
		panic(me.fail() + "argument " + arg.typed + " does not match parameter " + param.typed)
	}
	call.push(param)
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) *node {
	n := nodeInit("call")
	name := me.nameOfClassFunc(c.name, fn.name)
	if module == me.hmfile {
		n.value = name
	} else {
		n.value = module.name + "." + name
	}
	n.typed = fn.typed
	n.push(root)
	for ix, arg := range fn.args {
		if ix == 0 {
			continue
		}
		me.pushparam(n, arg)
	}
	return n
}

func (me *parser) call(module *hmfile) *node {
	token := me.token
	name := token.value
	fn := module.functions[name]
	args := fn.args
	me.eat("id")
	n := nodeInit("call")
	if module == me.hmfile {
		n.value = name
	} else {
		n.value = module.name + "." + name
	}
	n.typed = fn.typed
	for _, arg := range args {
		me.pushparam(n, arg)
	}
	return n
}

func (me *parser) eatvar(from *hmfile) *node {
	root := nodeInit("variable")
	localvarname := me.token.value
	if from == me.hmfile {
		root.value = localvarname
	} else {
		root.value = from.name + "." + localvarname
	}
	me.eat("id")
	for {
		if me.token.is == "." {
			if root.is == "variable" {
				sv := me.hmfile.getvar(root.value)
				if sv == nil {
					panic(me.fail() + "variable \"" + root.value + "\" out of scope")
				}
				root.typed = sv.typed
				root.is = "root-variable"
			}
			module, className := me.hmfile.moduleAndName(root.typed)
			rootClass, _ := module.getclass(className)
			if rootClass == nil {
				panic(me.fail() + "class \"" + root.typed + "\" does not exist")
			}
			me.eat(".")
			dotName := me.token.value
			me.eat("id")
			var member *node
			classVar, ok := rootClass.variables[dotName]
			if ok {
				fmt.Println("member variable \"" + dotName + "\" is type \"" + classVar.typed + "\"")
				member = nodeInit("member-variable")
				member.typed = classVar.typed
				member.value = dotName
				member.push(root)
			} else {
				nameOfFunc := me.nameOfClassFunc(rootClass.name, dotName)
				funcVar, ok := module.functions[nameOfFunc]
				if ok {
					fmt.Println("class function \"" + dotName + "\" returns \"" + funcVar.typed + "\"")
					member = me.callClassFunction(module, root, rootClass, funcVar)
				} else {
					panic(me.fail() + "class variable or function \"" + dotName + "\" does not exist")
				}
			}
			root = member
		} else if me.token.is == "[" {
			if root.is == "variable" {
				sv := me.hmfile.getvar(root.value)
				if sv == nil {
					panic(me.fail() + "variable out of scope")
				}
				root.typed = sv.typed
				root.is = "root-variable"
			}
			if !checkIsArray(root.typed) {
				panic(me.fail() + "root variable is not array")
			}
			atype := parseArrayType(root.typed)
			me.eat("[")
			member := nodeInit("array-member")
			index := me.calc()
			member.typed = atype
			member.push(index)
			member.push(root)
			root = member
			me.eat("]")
		} else {
			break
		}
	}
	if root.is == "variable" {
		if from == me.hmfile {
			sv := from.getvar(localvarname)
			if sv == nil {
				root.typed = "?"
			} else {
				root.typed = sv.typed
			}
		} else {
			sv := from.getstatic(localvarname)
			if sv == nil {
				panic(me.fail() + "static variable \"" + localvarname + "\" in module \"" + from.name + "\" not found")
			} else {
				root.typed = sv.typed
			}
		}
	}
	return root
}

func (me *parser) forceassign(malloc, mutable bool) *node {
	v := me.eatvar(me.hmfile)
	if !me.assignable(v) {
		panic(me.fail() + "expected variable for assignment but was \"" + v.typed + "\"")
	}
	return me.assign(v, malloc, mutable)
}

func (me *parser) assign(av *node, malloc, mutable bool) *node {
	op := me.token.is
	mustBeNumber := false
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" {
		mustBeNumber = true
	} else if op != "=" {
		panic(me.fail() + "unknown assign operation \"" + op + "\"")
	}
	me.eat(op)
	right := me.calc()
	if mustBeNumber && !isNumber(right.typed) {
		panic(me.fail() + "assign operation \"" + op + "\" requires number type")
	}
	if av.is == "variable" {
		sv := me.hmfile.getvar(av.value)
		if sv != nil {
			if !sv.mutable {
				panic(me.fail() + "variable \"" + sv.name + "\" is not mutable")
			}
		} else {
			if mustBeNumber {
				panic(me.fail() + "cannot operate \"" + op + "\" for variable \"" + av.value + "\" does not exist")
			} else {
				av.typed = right.typed
				if mutable {
					av.pushAttribute("mutable")
				}
				if !malloc {
					av.pushAttribute("no-malloc")
				}
				me.hmfile.scope.variables[av.value] = me.hmfile.varInit(right.typed, av.value, mutable, malloc)
			}
		}
	} else if av.is == "member-variable" || av.is == "array-member" {
		if av.typed != right.typed {
			panic(me.fail() + "member variable type " + av.typed + " does not match expression type " + right.typed)
		}
	} else {
		panic(me.fail() + "bad assignment \"" + av.is + "\"")
	}
	n := nodeInit(op)
	n.typed = "void"
	n.push(av)
	fmt.Println("assign set", av.string(0))
	n.push(right)
	return n
}

func (me *parser) allocClass(module *hmfile) *node {
	name := me.token.value
	me.eat("id")
	classDef, ok := module.classes[name]
	if !ok {
		panic(me.fail() + "class \"" + name + "\" does not exist")
	}
	gtypes := ""
	gsize := len(classDef.generics)
	if gsize > 0 {
		me.eat("[")
		gtypes += "["
		for i := 0; i < gsize; i++ {
			if i != 0 {
				me.eat("delim")
				gtypes += ","
			}
			gimpl := me.token.value
			me.eat("id")
			if _, ok := me.hmfile.types[gimpl]; !ok {
				panic(me.fail() + "generic implementation type \"" + gimpl + "\" does not exist")
			}
			gtypes += gimpl
		}
		me.eat("]")
		gtypes += "]"
	}

	n := nodeInit("new")
	if me.hmfile == module {
		n.typed = name + gtypes
	} else {
		n.typed = module.name + "." + name + gtypes
	}
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
		return me.allocClass(module)
	} else if module.getstatic(idname) != nil {
		fmt.Println("extern var")
		return me.eatvar(module)
	} else {
		panic(me.fail() + "extern " + extname + "." + idname + " does not exist")
	}
}

func (me *parser) block() *node {
	depth := me.token.depth
	block := nodeInit("block")
	for {
		done := false
		for me.token.is == "line" {
			me.eat("line")
			if me.token.is != "line" {
				if me.token.depth < depth {
					done = true
				}
				break
			}
		}
		if done {
			break
		}
		if me.token.is == "eof" {
			break
		}
		expr := me.expression()
		block.push(expr)
		if expr.is == "return" {
			fn := me.hmfile.scope.fn
			if fn.typed != expr.typed {
				panic(me.fail() + "function " + fn.name + " returns " + fn.typed + " but found " + expr.typed)
			}
			break
		}
	}
	fmt.Println("> block", block.string(0))
	return block
}

func (me *parser) continuing() *node {
	me.eat("continue")
	me.verify("line")
	n := nodeInit("continue")
	n.typed = "void"
	return n
}

func (me *parser) breaking() *node {
	me.eat("break")
	me.verify("line")
	n := nodeInit("break")
	n.typed = "void"
	return n
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
			n.push(me.getbool())
		} else {
			fmt.Println("> multi for")
			n.push(me.forceassign(true, true))
			me.eat("delim")
			n.push(me.getbool())
			me.eat("delim")
			n.push(me.forceassign(true, true))
		}
		me.eat("line")
	}
	n.push(me.block())
	return n
}

func (me *parser) ifexpr() *node {
	fmt.Println("> if")
	me.eat("if")
	n := nodeInit("if")
	n.typed = "void"
	n.push(me.getbool())
	me.eat("line")
	n.push(me.block())
	for me.token.is == "elif" {
		me.eat("elif")
		other := nodeInit("elif")
		other.push(me.getbool())
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
	fmt.Println("> comparison")
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
	right := me.calc()
	if left.typed != right.typed {
		panic(me.fail() + "types do not match")
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

func (me *parser) getbool() *node {
	n := me.calc()
	if n.typed != "bool" {
		panic(me.fail() + "must be boolean expression")
	}
	return n
}

func (me *parser) notbool() *node {
	me.eat("!")
	n := nodeInit("not")
	n.typed = "bool"
	n.push(me.getbool())
	fmt.Println("> not bool", n.string(0))
	return n
}

func (me *parser) binary(left *node, op string) *node {
	me.eat(op)
	right := me.term()
	if !isNumber(left.typed) || !isNumber(right.typed) {
		err := me.fail() + "binary operation must use numbers \"" + left.typed + "\" and \"" + right.typed + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	if left.typed != right.typed {
		err := me.fail() + "number types do not match \"" + left.typed + "\" and \"" + right.typed + "\""
		panic(err)
	}
	n := nodeInit(op)
	n.typed = left.typed
	n.push(left)
	n.push(right)
	return n
}

func (me *parser) initarray() *node {
	me.eat("[")
	size := me.calc()
	if size.typed != "int" {
		panic(me.fail() + "array size must be integer")
	}
	me.eat("]")
	n := nodeInit("array")
	typed := me.token.value
	me.eat("id")
	if _, ok := me.hmfile.imports[typed]; ok {
		me.eat(".")
		typed += "." + me.token.value
		me.eat("id")
		module, moduleType := me.hmfile.moduleAndName(typed)
		if _, ok := module.types[moduleType]; !ok {
			panic(me.fail() + "array type \"" + typed + "." + moduleType + "\" not found")
		}
	} else if _, ok := me.hmfile.types[typed]; !ok {
		panic(me.fail() + "array type \"" + typed + "\" not found")
	}
	n.typed = "[]" + typed
	n.push(size)
	fmt.Println("array node =", n.string(0))
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
			me.hmfile.program.compile(me.hmfile.program.out, path)
			fmt.Println("finished compiling " + name)
			fmt.Println("=== continue " + me.hmfile.name + " parse === ")
		}
		me.eat("line")
		if me.token.is == "line" || me.token.is == "eof" {
			break
		}
	}
}

func (me *parser) immutables() {
	me.eat("line")
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
		if me.token.is == "line" || me.token.is == "eof" {
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
		if me.token.is == "line" || me.token.is == "eof" {
			break
		}
	}
}

func (me *parser) defineClass() {
	me.eat("class")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.classes[name]; ok {
		panic(me.fail() + "class already defined")
	}
	me.eat("id")
	genericsMap := make(map[string]bool, 0)
	genericsOrder := make([]string, 0)
	if me.token.is == "[" {
		me.eat("[")
		for {
			gname := me.token.is
			me.eat("id")
			genericsMap[gname] = true
			genericsOrder = append(genericsOrder, gname)
			if me.token.is == "delim" {
				me.eat("delim")
				continue
			}
			if me.token.is == "]" {
				break
			}
			panic(me.fail() + "bad token \"" + me.token.is + "\" in class generic")
		}
		me.eat("]")
	}
	me.eat("line")
	memberorder := make([]string, 0)
	membermap := make(map[string]*variable)
	for {
		token := me.token
		if token.is == "line" {
			me.eat("line")
			break
		}
		if token.is == "eof" {
			break
		}
		if token.is == "id" {
			mname := token.value
			me.eat("id")
			if _, ok := membermap[mname]; ok {
				panic(me.fail() + "member name \"" + mname + "\" already used")
			}
			if _, ok := genericsMap[mname]; ok {
				panic(me.fail() + "cannot use \"" + mname + "\" as member name")
			}
			mtype := me.typedecl()
			me.eat("line")
			memberorder = append(memberorder, mname)
			membermap[mname] = me.hmfile.varInit(mtype, mname, true, true)
		}
	}
	me.hmfile.classOrder = append(me.hmfile.classOrder, name)
	me.hmfile.classes[name] = classInit(name, memberorder, membermap, genericsOrder)
	me.hmfile.types[name] = true
}

func (me *parser) calc() *node {
	node := me.term()
	for {
		token := me.token
		op := token.is
		if op == "and" || op == "or" {
			node = me.comparison(node, op)
			continue
		}
		if op == "=" || op == ">" || op == "<" || op == ">=" || op == "<=" || op == "!=" {
			node = me.comparison(node, op)
			continue
		}
		if op == "+" || op == "-" {
			node = me.binary(node, op)
			continue
		}
		break
	}
	return node
}

func (me *parser) term() *node {
	node := me.factor()
	for {
		token := me.token
		op := token.is
		if op == "*" || op == "/" {
			node = me.binary(node, op)
			continue
		}
		break
	}
	return node
}

func (me *parser) factor() *node {
	token := me.token
	op := token.is
	if _, ok := primitives[op]; ok {
		me.eat(op)
		n := nodeInit(op)
		n.typed = op
		n.value = token.value
		return n
	}
	if op == "id" {
		name := token.value
		if _, ok := me.hmfile.functions[name]; ok {
			return me.call(me.hmfile)
		}
		if _, ok := me.hmfile.types[name]; ok {
			if _, ok := me.hmfile.classes[name]; ok {
				return me.allocClass(me.hmfile)
			}
			panic(me.fail() + "bad type \"" + name + "\" definition")
		}
		if _, ok := me.hmfile.imports[name]; ok {
			return me.extern()
		}
		if me.hmfile.getvar(name) == nil {
			panic(me.fail() + "variable out of scope")
		}
		return me.eatvar(me.hmfile)
	}
	if op == "[" {
		return me.initarray()
	}
	if op == "(" {
		me.eat("(")
		n := me.calc()
		n.pushAttribute("parenthesis")
		me.eat(")")
		return n
	}
	if op == "!" {
		return me.notbool()
	}
	panic(me.fail() + "unknown factor")
}
