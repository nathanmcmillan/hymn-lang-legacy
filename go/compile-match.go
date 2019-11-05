package main

import (
	"strings"
)

func (me *cfile) compileIs(n *node) *codeblock {
	code := ""
	code += me.walrusMatch(n)
	using := n.has[0]
	match := me.eval(using)

	if match.data().checkIsSomeOrNone() {
		caseOf := n.has[1]
		if caseOf.is == "some" {
			code += match.code() + " != NULL"
		} else if caseOf.is == "none" {
			code += match.code() + " == NULL"
		}
		return codeBlockOne(n, code)
	}

	baseEnum, _, _ := using.data().checkIsEnum()
	if baseEnum.simple {
		code += match.code()
	} else {
		code += using.idata.name + "->type"
	}

	code += " == "

	caseOf := n.has[1]
	if caseOf.is == "match-enum" {
		matchBaseEnum, matchBaseUn, _ := caseOf.data().checkIsEnum()
		enumNameSpace := using.data().module.enumNameSpace(matchBaseEnum.name)
		code += using.data().module.enumTypeName(enumNameSpace, matchBaseUn.name)
	} else {
		compare := me.eval(caseOf)
		compareEnum, _, _ := compare.data().checkIsEnum()
		code += compare.code()
		if !compareEnum.simple {
			code += "->type"
		}
	}

	return codeBlockOne(n, code)
}

func (me *cfile) compileMatch(n *node) *codeblock {
	code := ""
	code += me.walrusMatch(n)
	using := n.has[0]
	match := me.eval(using)

	if match.data().checkIsSomeOrNone() {
		return me.compileMatchNull(match, n, code)
	}

	test := match.code()
	isEnum := false
	var enumNameSpace string

	if using.is == "variable" {
		if baseEnum, _, ok := using.data().checkIsEnum(); ok {
			isEnum = true
			enumNameSpace = using.data().module.enumNameSpace(baseEnum.name)
			if !baseEnum.simple {
				test = using.idata.name + "->type"
			}
		}
	}

	code += "switch (" + test + ") {\n"
	ix := 1
	size := len(n.has)
	for ix < size {
		caseOf := n.has[ix]
		thenDo := n.has[ix+1]
		if caseOf.is == "_" {
			code += fmc(me.depth) + "default: {\n"
		} else {
			if isEnum {
				code += fmc(me.depth) + "case " + using.data().module.enumTypeName(enumNameSpace, caseOf.is) + ": {\n"
				tempname := ""
				if len(caseOf.has) > 0 {
					temp := caseOf.has[0]
					tempname = temp.idata.name
					tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
					me.scope.variables[tempname] = tempv
					ref := me.eval(n.has[0]).code()
					code += fmc(me.depth + 1)
					code += fmtassignspace(temp.data().typeSig()) + tempname + " = " + ref + ";\n"
				}
			} else {
				code += fmc(me.depth) + "case " + caseOf.is + ":\n"
			}
		}
		thenBlock := me.eval(thenDo).code()
		me.depth++
		if thenBlock != "" {
			code += me.maybeFmc(thenBlock, me.depth) + thenBlock + me.maybeColon(thenBlock) + "\n"
		}
		if !strings.Contains(thenBlock, "return") {
			code += fmc(me.depth) + "break;"
		}
		code += " }\n"
		me.depth--
		ix += 2
	}
	code += fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) compileMatchNull(match *codeblock, n *node, code string) *codeblock {
	ifNull := ""
	ifNotNull := ""
	ix := 1
	size := len(n.has)
	for ix < size {
		block := ""
		tempname := ""
		caseOf := n.has[ix]
		if caseOf.is == "some" {
			if len(caseOf.has) > 0 {
				temp := caseOf.has[0]
				tempname = temp.idata.name
				tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
				me.scope.variables[tempname] = tempv
				ref := me.eval(n.has[0]).code()
				block += fmtassignspace(temp.data().typeSig()) + tempname + " = " + ref + ";\n"
			}
		}
		c := me.eval(n.has[ix+1]).code()
		block += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		if tempname != "" {
			delete(me.scope.variables, tempname)
		}
		if caseOf.is == "some" {
			ifNotNull = block
		} else if caseOf.is == "none" {
			ifNull = block
		}
		ix += 2
	}

	if ifNull != "" && ifNotNull != "" {
		code += "if (" + match.code() + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "} else {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	} else if ifNull != "" {
		code += "if (" + match.code() + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "}"

	} else {
		code += "if (" + match.code() + " != NULL) {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	}

	return codeBlockOne(n, code)
}
