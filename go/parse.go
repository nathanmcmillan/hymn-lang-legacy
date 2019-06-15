package main

import (
	"fmt"
	"strings"
)

func (me *variable) string() string {
	return "{is:" + me.is + ", name:" + me.name + "}"
}

func (me *program) dump() string {
	s := ""
	lv := 0
	if len(me.classOrder) > 0 {
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
	}
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

func (me *parser) next() {
	me.pos++
	me.token = me.tokens[me.pos]
}

func (me *parser) peek() *token {
	return me.tokens[me.pos+1]
}

func (me *parser) fail() string {
	return fmt.Sprintf("token %s at position %d\n", me.tokens[me.pos].string(), me.pos)
}

func (me *parser) skipLines() {
	for me.pos != len(me.tokens) {
		token := me.token
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
	me.skipLines()
	for me.token.is != "eof" {
		me.expression()
		if me.token.is == "line" {
			me.eat("line")
		}
	}
	delete(me.program.functions, "echo")
	return me.program
}

func (me *parser) verify(want string) {
	token := me.token
	if token.is != want {
		panic(fmt.Sprintf("unexpected token was "+token.string()+" instead of {type:"+want+"} on line %d", (me.pos + 1)))
	}
}

func (me *parser) eat(want string) {
	me.verify(want)
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
			me.verify("line")
		}
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
	} else if op == "line" {
		return nil
	} else if op == "eof" {
		return nil
	}
	panic("unknown expression " + me.fail())
}

func (me *parser) maybeIgnore(depth int) {
	for {
		if me.token.is == "line" {
			me.eat("line")
			break
		}
	}
	for me.pos != len(me.tokens) {
		token := me.token
		if token.is != "line" {
			break
		}
		me.next()
	}
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
	fn.name = name
	fn.typed = "void"
	for {
		if me.token.is == "line" {
			me.next()
			break
		}
		if me.token.is == "return-type" {
			me.next()
			fn.typed = me.token.value
			me.eat("id")
			continue
		}
		if me.token.is == "id" {
			typed := me.token.value
			me.eat("id")
			me.eat(":")
			arg := me.token.value
			me.eat("id")
			fn.args = append(fn.args, varInit(typed, arg))
			continue
		}
		panic("unexpected token in function definition " + me.fail())
	}
	me.program.pushScope()
	me.program.scope.fn = fn
	for _, arg := range fn.args {
		me.program.scope.variables[arg.name] = arg
	}
	for {
		token = me.token
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
	me.verify("line")
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
		ca := me.factor()
		if ca.typed != arg.is && arg.is != "?" {
			panic("argument " + arg.is + " does not match parameter " + ca.typed + " " + me.fail())
		}
		n.push(ca)
	}
	return n
}

func parseArrayType(typed string) string {
	return typed[0:strings.Index(typed, "[")]
}

func checkIsArray(typed string) bool {
	return strings.HasSuffix(typed, "[]")
}

func (me *parser) eatvar() *node {
	root := nodeInit("variable")
	root.value = me.token.value
	me.eat("id")
	for {
		if me.token.is == "." {
			if root.is == "variable" {
				sv := me.program.scope.getVar(root.value)
				if sv == nil {
					panic("variable out of scope " + me.fail())
				}
				root.typed = sv.is
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
		} else if me.token.is == "[" {
			if root.is == "variable" {
				sv := me.program.scope.getVar(root.value)
				if sv == nil {
					panic("variable out of scope " + me.fail())
				}
				root.typed = sv.is
				root.is = "root-variable"
			}
			if !checkIsArray(root.typed) {
				panic("root variable is not array " + me.fail())
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
		sv := me.program.scope.getVar(root.value)
		if sv == nil {
			root.typed = "?"
		} else {
			root.typed = sv.is
		}
	}
	return root
}

func (me *parser) assign() *node {
	assignVar := me.eatvar()
	op := me.token.is
	mustBeNumber := false
	if op == "=" {
	} else if op == "+=" || op == "-=" || op == "*=" || op == "/=" {
		mustBeNumber = true
	} else {
		panic("unknown assign operation " + me.fail())
	}
	me.eat(op)
	calc := me.calc()
	if mustBeNumber && !isNumber(calc.typed) {
		panic("assign operation " + op + " requires number type")
	}
	if assignVar.is == "variable" {
		assignVar.typed = calc.typed
		me.program.scope.variables[assignVar.value] = varInit(calc.typed, assignVar.value)
		// TODO mutable vs immutable
		// TODO if mustBeNumber than also must exist already and not be set any more
	} else if assignVar.is == "member-variable" {
		if assignVar.typed != calc.typed {
			panic("member variable type " + assignVar.typed + " does not match expression type " + calc.typed + " " + me.fail())
		}
	}
	n := nodeInit(op)
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

func (me *parser) enclosing() *node {
	depth := me.token.depth
	fmt.Println("> enclose depth", depth)
	enclose := nodeInit("scope")
	me.program.pushScope()
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
		enclose.push(expr)
		if expr.is == "return" {
			fn := me.program.scope.fn
			if fn.typed != expr.typed {
				panic("function " + fn.name + " returns " + fn.typed + " but found " + expr.typed)
			}
			break
		}
	}
	me.program.popScope()
	fmt.Println("> enclose", enclose.string(0))
	return enclose
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

func (me *parser) forexpr() *node {
	fmt.Println("> for")
	me.eat("for")
	n := nodeInit("for")
	n.typed = "void"
	me.eat("line")
	n.push(me.enclosing())
	return n
}

func (me *parser) ifexpr() *node {
	fmt.Println("> if")
	me.eat("if")
	n := nodeInit("if")
	n.typed = "void"
	n.push(me.boolexpr())
	me.eat("line")
	n.push(me.enclosing())
	for me.token.is == "elif" {
		me.eat("elif")
		other := nodeInit("elif")
		other.push(me.boolexpr())
		me.eat("line")
		other.push(me.enclosing())
		n.push(other)
	}
	if me.token.is == "else" {
		me.eat("else")
		me.eat("line")
		n.push(me.enclosing())
	}
	return n
}

func (me *parser) boolexpr() *node {
	fmt.Println("> boolexpr")
	left := me.calc()
	var typed string
	if me.token.is == "=" {
		typed = "equal"
		me.eat("=")
	} else if me.token.is == ">" || me.token.is == ">=" || me.token.is == "<" || me.token.is == "<=" {
		if !isNumber(left.typed) {
			panic("left side of comparison must be a number " + me.fail())
		}
		typed = me.token.is
		me.eat(me.token.is)
	} else {
		if left.typed == "bool" {
			typed = "boolexpr"
			n := nodeInit(typed)
			n.typed = "bool"
			n.push(left)
			fmt.Println("> bool using", typed)
			fmt.Println("> just", left.string(0))
			return n
		}
		panic("unknown token for boolean expression " + me.fail())
	}
	right := me.calc()
	if left.typed != right.typed {
		panic("left and right side of comparison must match " + me.fail())
	}
	n := nodeInit(typed)
	n.typed = "bool"
	n.push(left)
	n.push(right)
	fmt.Println("> bool using", typed)
	fmt.Println("> left", left.string(0))
	fmt.Println("> right", right.string(0))
	return n
}

func (me *parser) binary(left, right *node, op string) *node {
	if !isNumber(left.typed) || !isNumber(right.typed) {
		panic("binary operation must use integers")
	}
	if left.typed != right.typed {
		panic("number types do not match")
	}
	n := nodeInit(op)
	n.typed = left.typed
	n.push(left)
	n.push(right)
	return n
}

func (me *parser) array(is string) *node {
	me.eat("[")
	size := me.calc()
	if size.typed != "int" {
		panic("array size must be integer")
	}
	me.eat("]")
	n := nodeInit("array")
	n.typed = is + "[]"
	n.push(size)
	fmt.Println("array node =", n.string(0))
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
	if _, ok := me.program.primitives[op]; ok {
		me.eat(op)
		n := nodeInit(op)
		n.typed = op
		n.value = token.value
		return n
	}
	if op == "id" {
		name := token.value
		if _, ok := me.program.functions[name]; ok {
			return me.call()
		}
		if _, ok := me.program.types[name]; ok {
			if me.peek().is == "[" {
				me.eat(op)
				return me.array(name)
			}
			panic("bad array definition " + me.fail())
		}
		if me.program.scope.getVar(name) == nil {
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
	me.program.types[name] = true
}
