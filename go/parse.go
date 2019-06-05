package main

import "fmt"

func (me *variable) string() string {
	return "{is:" + me.is + ", name:" + me.name + "}"
}

func (me *program) dump() string {
	s := ""
	lv := 0
	s += fmc(lv) + "functions{\n"
	for name, function := range me.functions {
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
	s += fmc(lv) + "}\n"
	return s
}

func (me *program) libInit() {
	e := funcInit()
	e.typed = "void"
	e.args = append(e.args, varInit("string", "s"))
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
	for me.token.is != "line" {
		arg := me.token.value
		fn.args = append(fn.args, varInit("int", arg))
		me.eat("id")
	}
	me.eat("line")
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
			fn.typed = expr.typed
			break
		}
	}
	me.program.popScope()
	program.functions[name] = fn
	program.functionOrder = append(program.functionOrder, name)
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
	for range args {
		n.push(me.calc())
	}
	return n
}

func (me *parser) assign() *node {
	token := me.token
	me.eat("id")
	me.eat("=")
	calc := me.calc()
	n := nodeInit("assign")
	n.value = token.value
	n.typed = calc.typed
	n.push(calc)
	me.program.scope.variables[n.value] = varInit(n.typed, n.value)
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
	n.value = name
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
		sv, ok := me.program.scope.variables[name]
		if !ok {
			panic("variable out of scope")
		}
		me.eat(op)
		n := nodeInit("id")
		n.typed = sv.is
		n.value = token.value
		return n
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
	vs := make([]*variable, 0)
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
			vs = append(vs, varInit(vt, vn))
		}
	}
	me.program.classes[name] = vs
}
