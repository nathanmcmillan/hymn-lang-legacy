package main

import "strconv"

func (me *cfile) hintEval(n *node, hint *varData) *cnode {
	op := n.is
	if op == "=" || op == ":=" {
		code := me.assingment(n)
		return codeNode(n, code)
	}
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" || op == "%=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		code := me.assignmentUpdate(n)
		return codeNode(n, code)
	}
	if op == "new" {
		return me.allocClass(n)
	}
	if op == "enum" {
		code := me.allocEnum(n)
		return codeNode(n, code)
	}
	if op == "cast" {
		return me.compileCast(n)
	}
	if op == "concat" {
		size := len(n.has)
		code := ""
		if size == 2 {
			code += "hmlib_concat("
			code += me.eval(n.has[0]).code
			code += ", "
			code += me.eval(n.has[1]).code
		} else {
			code += "hmlib_concat_varg(" + strconv.Itoa(size)
			for _, snode := range n.has {
				code += ", " + me.eval(snode).code
			}
		}
		code += ")"
		return codeNode(n, code)
	}
	if op == "+sign" {
		return me.compilePrefixPos(n)
	}
	if op == "-sign" {
		return me.compilePrefixNeg(n)
	}
	if op == "+" || op == "-" || op == "*" || op == "/" || op == "%" || op == "&" || op == "|" || op == "^" || op == "<<" || op == ">>" {
		return me.compileBinaryOp(n)
	}
	if op == "fn-ptr" {
		return me.compileFnPtr(n, hint)
	}
	if op == "variable" {
		return me.compileVariable(n, hint)
	}
	if op == "root-variable" {
		v := me.getvar(n.idata.name)
		return codeNode(n, v.cName)
	}
	if op == "array-member" {
		index := me.eval(n.has[0])
		root := me.eval(n.has[1])
		code := root.code + "[" + index.code + "]"
		return codeNode(n, code)
	}
	if op == "member-variable" {
		return me.compileMemberVariable(n)
	}
	if op == "tuple-index" {
		return me.compileTupleIndex(n)
	}
	if op == "call" {
		return me.compileCall(n)
	}
	if op == "array" {
		code := me.allocArray(n)
		return codeNode(n, code)
	}
	if op == "slice" {
		code := me.allocSlice(n)
		return codeNode(n, code)
	}
	if op == "return" {
		in := me.eval(n.has[0])
		code := "return " + in.code
		cn := codeNode(n, code)
		cn.push(in)
		return cn
	}
	if op == "boolexpr" {
		code := me.eval(n.has[0]).code
		return codeNode(n, code)
	}
	if op == "equal" {
		return me.compileEqual(n)
	}
	if op == "not" {
		code := "!" + me.eval(n.has[0]).code
		return codeNode(n, code)
	}
	if op == "not-equal" {
		code := me.eval(n.has[0]).code
		code += " != "
		code += me.eval(n.has[1]).code
		return codeNode(n, code)
	}
	if op == ">" || op == ">=" || op == "<" || op == "<=" {
		code := me.eval(n.has[0]).code
		code += " " + op + " "
		code += me.eval(n.has[1]).code
		return codeNode(n, code)
	}
	if op == "?" {
		return me.compileTernary(n)
	}
	if op == "and" || op == "or" {
		return me.compileAndOr(n)
	}
	if op == "block" {
		return me.block(n)
	}
	if op == "break" {
		return codeNode(n, "break")
	}
	if op == "continue" {
		return codeNode(n, "continue")
	}
	if op == "goto" {
		return codeNode(n, "goto "+n.value)
	}
	if op == "label" {
		return codeNode(n, n.value+":")
	}
	if op == "pass" {
		return codeNode(n, "")
	}
	if op == "match" {
		return me.compileMatch(n)
	}
	if op == "is" {
		return me.compileIs(n)
	}
	if op == "for" {
		return me.compileFor(n)
	}
	if op == "if" {
		return me.compileIf(n)
	}
	if op == TokenRawString {
		return me.compileRawString(n)
	}
	if op == TokenString {
		return me.compileString(n)
	}
	if op == TokenChar {
		return me.compileChar(n)
	}
	if op == "none" {
		return me.compileNone(n)
	}
	if _, ok := primitives[op]; ok {
		return codeNode(n, n.value)
	}
	panic("eval unknown operation " + n.string(0))
}
