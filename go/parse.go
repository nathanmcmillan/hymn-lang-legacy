package main

import "fmt"

type node struct {
	is    string
	value string
	has   []*node
}

type variable struct {
	is   string
	name string
}

type scope struct {
	root      *scope
	variables map[string]*variable
}

type function struct {
	args        []*variable
	expressions []*node
}

type program struct {
	imports       map[string]bool
	objects       map[string][]*variable
	rootScope     *scope
	scope         *scope
	functions     map[string]*function
	functionOrder []string
}

type parser struct {
	tokens  []*token
	token   *token
	pos     int
	program *program
}

func (me *variable) string() string {
	return "{is:" + me.is + ", name:" + me.name + "}"
}

func fmc(level int, content string) string {
	for i := 0; i < level; i++ {
		content = "  " + content
	}
	return content
}

func (me *node) string(lv int) string {
	s := ""
	s += fmc(lv, "{is:"+me.is)
	if me.value != "" {
		s += ", value:" + me.value
	}
	if len(me.has) > 0 {
		s += ", has[\n"
		lv++
		for ix, has := range me.has {
			if ix > 0 {
				s += ",\n"
			}
			s += has.string(lv)
		}
		lv--
		s += "\n"
		s += fmc(lv, "]")
	}
	s += "}"
	return s
}

func (me *program) dump() string {
	s := ""
	lv := 0
	s += fmc(lv, "functions:{\n")
	for name, function := range me.functions {
		lv++
		s += fmc(lv, name+":{\n")
		lv++
		s += fmc(lv, "args:[\n")
		lv++
		for ix, arg := range function.args {
			if ix > 0 {
				s += ",\n"
			}
			s += fmc(lv, arg.string())
		}
		lv--
		s += "\n"
		s += fmc(lv, "]\n")
		s += fmc(lv, "expressions:[\n")
		lv++
		for ix, expr := range function.expressions {
			if ix > 0 {
				s += ",\n"
			}
			s += expr.string(lv)
		}
		lv--
		s += "\n"
		s += fmc(lv, "]\n")
		lv--
		s += fmc(lv, "}\n")
		lv--
	}
	s += fmc(lv, "}\n")
	return s
}

func nodeInit(is string) *node {
	n := &node{}
	n.is = is
	n.has = make([]*node, 0)
	return n
}

func (me *node) push(n *node) {
	me.has = append(me.has, n)
}

func scopeInit(root *scope) *scope {
	s := &scope{}
	s.root = root
	s.variables = make(map[string]*variable)
	return s
}

func programInit() *program {
	p := &program{}
	p.imports = make(map[string]bool)
	p.rootScope = scopeInit(nil)
	p.scope = p.rootScope
	p.functions = make(map[string]*function)
	p.functionOrder = make([]string, 0)
	p.libInit()
	return p
}

func (me *program) libInit() {
	e := funcInit()
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
		me.statement()
	}
	return me.program
}

func (me *parser) eat(want string) {
	token := me.token
	if token.is != want {
		panic(fmt.Sprintf("unexpected token was "+token.string()+" instead of {type:"+want+"} at position %d", me.pos))
	}
	me.next()
}

func (me *parser) statement() *node {
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
	} else if op == "object" {
		me.object()
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
	panic("unknown statement " + me.fail())
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
	for me.token.is != "line" {
		fn.args = append(fn.args, varInit("int", me.token.value))
		me.eat("id")
	}
	me.eat("line")
	for {
		token = me.token
		if token.is == "line" {
			me.eat("line")
			break
		}
		if token.is == "eof" {
			break
		}
		stat := me.statement()
		fn.expressions = append(fn.expressions, stat)
		if stat.is == "return" {
			break
		}
	}
	program.functions[name] = fn
	program.functionOrder = append(program.functionOrder, name)
}

func (me *parser) returning() *node {
	me.eat("return")
	n := nodeInit("return")
	n.push(me.calc())
	return n
}

func (me *parser) call() *node {
	token := me.token
	name := token.value
	args := me.program.functions[name].args
	me.eat("id")
	n := nodeInit("call")
	n.value = name
	for range args {
		n.push(me.calc())
	}
	return n
}

func (me *parser) assign() *node {
	token := me.token
	me.eat("id")
	me.eat("=")
	n := nodeInit("assign")
	n.value = token.value
	n.push(me.calc())
	return n
}

func (me *parser) construct() *node {
	me.eat("new")
	token := me.token
	me.eat("id")
	name := token.value
	if _, ok := me.program.objects[name]; !ok {
		panic("object does not exist " + me.fail())
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
	n := nodeInit(op)
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
	if op == "number" {
		me.eat(op)
		n := nodeInit("number")
		n.value = token.value
		return n
	}
	if op == "string" {
		me.eat(op)
		n := nodeInit("string")
		n.value = token.value
	}
	if op == "id" {
		if _, ok := me.program.functions[token.value]; ok {
			return me.call()
		}
		me.eat(op)
		n := nodeInit("id")
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

func (me *parser) object() {
	me.eat("object")
	token := me.token
	name := token.value
	if _, ok := me.program.objects[name]; ok {
		panic("object already defined " + me.fail())
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
			vname := token.value
			vs = append(vs, varInit("int", vname))
			me.eat("id")
			me.eat("line")
			break
		}
	}
	me.program.objects[name] = vs
}
