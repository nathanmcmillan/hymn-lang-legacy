package main

import "fmt"

func (me *variable) string() string {
	return "{is:" + me.is + ", name:" + me.name + "}"
}

func (me *program) dump() string {
	s := ""
	lv := 0
	s += fmc(lv) + "classes[\n"
	for _, name := range me.classOrder {
		class := me.classes[name]
		lv++
		s += fmc(lv) + class.name + "[\n"
		lv++
		for _, classVar := range class.variables {
			s += fmc(lv) + "{name:" + classVar.name + ", is:" + classVar.is + "}\n"
		}
		lv--
		s += fmc(lv) + "]\n"
		lv--
	}
	s += fmc(lv) + "]\n"
	s += fmc(lv) + "functions[\n"
	for _, name := range me.functionOrder {
		function := me.functions[name]
		lv++
		s += fmc(lv) + name + "{\n"
		lv++
		if len(function.args) > 0 {
			s += fmc(lv) + "args[\n"
			lv++
			for _, arg := range function.args {
				s += fmc(lv) + arg.string() + "\n"
			}
			lv--
			s += fmc(lv) + "]\n"
		}
		if len(function.expressions) > 0 {
			s += fmc(lv) + "expressions[\n"
			lv++
			for _, expr := range function.expressions {
				s += expr.string(lv) + "\n"
			}
			lv--
			s += fmc(lv) + "]\n"
		}
		lv--
		s += fmc(lv) + "}\n"
		lv--
	}
	s += fmc(lv) + "]\n"
	return s
}

func (me *program) libInit() {
	e := funcInit()
	e.typed = "void"
	e.args = append(e.args, varInit("?", "s"))
	me.functions["echo"] = e
}

func funcInit() *function {
	f := &function{}
	f.args = make([]*variable, 0)
	f.expressions = make([]*node, 0)
	return f
}

func varInit(is, name string) *variable {
	v := &variable{}
	v.is = is
	v.name = name
	return v
}

func (me *parser) next() {
	me.pos++
	me.token = me.tokens[me.pos]
}

func (me *parser) peek() *token {
	return me.tokens[me.pos]
}

func (me *parser) fail() string {
	return fmt.Sprintf("token %s at position %d\n", me.tokens[me.pos].string(), me.pos)
}

func (me *parser) forLines() {
	for me.pos != len(me.tokens) {
		token := me.peek()
		if token.is != "line" {
			break
		}
		me.next()
	}
}

func parse(tokens []*token) *program {
	me := parser{}
	me.tokens = tokens
	me.token = tokens[0]
	me.program = programInit()
	me.forLines()
	for me.token.is != "eof" {
		me.expression()
	}
	delete(me.program.functions, "echo")
	return me.program
}

func (me *parser) eat(want string) {
	token := me.token
	if token.is != want {
		panic(fmt.Sprintf("unexpected token was "+token.string()+" instead of {type:"+want+"} on line %d", (me.pos + 1)))
	}
	me.next()
}

func (me *parser) expression() *node {
	token := me.token
	op := token.is
	if op == "id" {
		name := token.value
		var n *node
		if _, ok := me.program.functions[name]; ok {
			n = me.call()
		} else {
			n = me.assign()
		}
		me.eat("line")
		return n
	} else if op == "function" {
		me.function()
		return nil
	} else if op == "new" {
		me.construct()
		return nil
	} else if op == "class" {
		me.class()
		return nil
	} else if op == "free" {
		me.free()
		return nil
	} else if op == "return" {
		n := me.returning()
		me.eat("line")
		return n
	} else if op == "line" {
		me.eat("line")
		return nil
	} else if op == "eof" {
		return nil
	}
	panic("unknown expression " + me.fail())
}

func (me *parser) function() {
	program := me.program
	me.eat("function")
	token := me.token
	name := token.value
	if _, ok := program.functions[name]; ok {
		panic("function already defined " + me.fail())
	}
	me.eat("id")
	fn := funcInit()
	fn.typed = "void"
	for true {
		if me.token.is == "line" {
			me.eat("line")
			break
		}
		if me.token.is == ":" {
			me.eat(":")
			fn.typed = me.token.value
			me.eat("id")
			continue
		}
		if me.token.is == "id" {
			arg := me.token.value
			me.eat("id")
			me.eat("=")
			typed := me.token.value
			me.eat("id")
			fn.args = append(fn.args, varInit(typed, arg))
			continue
		}
		panic("unexpected token in function definition " + me.fail())
	}
	me.program.pushScope()
	for _, arg := range fn.args {
		me.program.scope.variables[arg.name] = arg
	}
	for {
		token = me.token
		if token.is == "line" {
			me.eat("line")
			break
		}
		if token.is == "eof" {
			break
		}
		expr := me.expression()
		fn.expressions = append(fn.expressions, expr)
		if expr.is == "return" {
			if fn.typed != expr.typed {
				panic("function " + name + " returns " + fn.typed + " but found " + expr.typed)
			}
			break
		}
	}
	me.program.popScope()
	program.functionOrder = append(program.functionOrder, name)
	program.functions[name] = fn
}

func (me *parser) returning() *node {
	me.eat("return")
	calc := me.calc()
	n := nodeInit("return")
	n.typed = calc.typed
	n.push(calc)
	return n
}

func (me *parser) call() *node {
	token := me.token
	name := token.value
	fn := me.program.functions[name]
	args := fn.args
	me.eat("id")
	n := nodeInit("call")
	n.value = name
	n.typed = fn.typed
	for _, arg := range args {
		ca := me.calc()
		if ca.typed != arg.is && arg.is != "?" {
			panic("argument " + arg.is + " does not match parameter " + ca.typed + " " + me.fail())
		}
		n.push(ca)
	}
	return n
}

func (me *parser) eatvar() *node {
	root := nodeInit("variable")
	root.value = me.token.value
	me.eat("id")
	for me.token.is == "." {
		fmt.Println("root.value =", root.value)
		if root.is == "variable" {
			scopeVar, ok := me.program.scope.variables[root.value]
			if !ok {
				panic("variable out of scope " + me.fail())
			}
			root.typed = scopeVar.is
			root.is = "root-variable"

		}
		rootClass, ok := me.program.classes[root.typed]
		if !ok {
			panic("class " + root.typed + " does not exist " + me.fail())
		}
		me.eat(".")
		member := nodeInit("member-variable")
		memberName := me.token.value
		member.value = memberName
		classVar, ok := rootClass.variables[memberName]
		if !ok {
			panic("member variable " + memberName + " does not exist " + me.fail())
		}
		member.typed = classVar.is
		member.push(root)
		root = member
		me.eat("id")
	}
	if root.is == "variable" {
		scopeVar, ok := me.program.scope.variables[root.value]
		if ok {
			root.typed = scopeVar.is
		} else {
			root.typed = "?"
		}
	}
	return root
}

func (me *parser) assign() *node {
	assignVar := me.eatvar()
	me.eat("=")
	calc := me.calc()
	if assignVar.is == "variable" {
		assignVar.typed = calc.typed
		me.program.scope.variables[assignVar.value] = varInit(calc.typed, assignVar.value)
	} else if assignVar.is == "member-variable" {
		if assignVar.typed != calc.typed {
			panic("member variable type " + assignVar.typed + " does not match expression type " + calc.typed + " " + me.fail())
		}
	}
	n := nodeInit("assign")
	n.typed = "void"
	n.push(assignVar)
	fmt.Println("assign set", assignVar.string(0))
	n.push(calc)
	return n
}

func (me *parser) construct() *node {
	me.eat("new")
	token := me.token
	me.eat("id")
	name := token.value
	if _, ok := me.program.classes[name]; !ok {
		panic("class does not exist " + me.fail())
	}
	n := nodeInit("new")
	n.typed = name
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

func (me *parser) binary(left, right *node, op string) *node {
	if left.typed != "int" || right.typed != "int" {
		panic("binary operation must use integers")
	}
	n := nodeInit(op)
	n.typed = "int"
	n.push(left)
	n.push(right)
	return n
}

func (me *parser) calc() *node {
	node := me.term()
	for true {
		token := me.token
		op := token.is
		if op == "+" || op == "-" {
			me.eat(op)
			node = me.binary(node, me.term(), op)
			continue
		}
		break
	}
	return node
}

func (me *parser) term() *node {
	node := me.factor()
	for true {
		token := me.token
		op := token.is
		if op == "*" || op == "/" {
			me.eat(op)
			node = me.binary(node, me.term(), op)
			continue
		}
		break
	}
	return node
}

func (me *parser) factor() *node {
	token := me.token
	op := token.is
	if op == "int" {
		me.eat(op)
		n := nodeInit("int")
		n.typed = "int"
		n.value = token.value
		return n
	}
	if op == "string" {
		me.eat(op)
		n := nodeInit("string")
		n.typed = "string"
		n.value = token.value
		return n
	}
	if op == "id" {
		name := token.value
		if _, ok := me.program.functions[name]; ok {
			return me.call()
		}
		_, ok := me.program.scope.variables[name]
		if !ok {
			panic("variable out of scope " + me.fail())
		}
		return me.eatvar()
	}
	if op == "new" {
		return me.construct()
	}
	if op == "(" {
		me.eat("(")
		n := me.calc()
		me.eat(")")
		return n
	}
	panic("unknown factor " + me.fail())
}

func (me *parser) class() {
	me.eat("class")
	token := me.token
	name := token.value
	if _, ok := me.program.classes[name]; ok {
		panic("class already defined " + me.fail())
	}
	me.eat("id")
	me.eat("line")
	vorder := make([]string, 0)
	vmap := make(map[string]*variable)
	for true {
		token := me.token
		if token.is == "line" {
			me.eat("line")
			break
		}
		if token.is == "eof" {
			break
		}
		if token.is == "id" {
			vn := token.value
			me.eat("id")
			token := me.token
			me.eat("id")
			vt := token.value
			me.eat("line")
			vorder = append(vorder, vn)
			vmap[vn] = varInit(vt, vn)
		}
	}
	me.program.classOrder = append(me.program.classOrder, name)
	me.program.classes[name] = classInit(name, vorder, vmap)
}
