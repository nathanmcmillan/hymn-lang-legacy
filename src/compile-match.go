package main

import (
	"strings"
)

func (me *cfile) compileIsSomeOrNone(code string, n, caseOf *node, match *codeblock) *codeblock {
	if caseOf.is == "some" {
		if len(caseOf.has) > 0 {
			temphas := caseOf.has[0]
			idata := temphas.idata.name
			tempname := "match_" + me.temp()
			tempv := temphas.data().getnamedvariable(tempname, false)
			me.scope.renaming[idata] = tempname
			me.scope.variables[tempname] = tempv
			prepend := fmtassignspace(match.data().typeSig(me)) + tempname + ";\n" + fmc(me.depth)
			code := "(" + tempname + " = " + match.code() + ") != NULL"
			cb := &codeblock{}
			cb.prepend(codeBlockOne(n, prepend))
			cb.current = codeNode(n, code)
			return cb
		}
		code += match.code() + " != NULL"

	} else if caseOf.is == "none" {
		code += match.code() + " == NULL"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileIs(n *node) *codeblock {
	code := ""
	using := n.has[0]
	caseOf := n.has[1]
	match := me.eval(using)

	if match.data().isSomeOrNone() {
		return me.compileIsSomeOrNone(code, n, caseOf, match)
	}

	cb := &codeblock{}

	tempname := ""

	if using.is == "variable" {
		name := me.getvar(using.idata.name).cname
		tempname = name
	}

	assignment := ""
	if len(caseOf.has) > 0 {
		temphas := caseOf.has[0]
		idata := temphas.idata.name
		if tempname == "" {
			tempname = "match_" + me.temp()
			tempv := temphas.data().getnamedvariable(tempname, false)
			me.scope.variables[tempname] = tempv
			prepend := fmtassignspace(match.data().typeSig(me)) + tempname + ";\n" + fmc(me.depth)
			assignment = "(" + tempname + " = " + match.code() + ")"
			cb.prepend(codeBlockOne(n, prepend))
		}
		me.scope.renaming[idata] = tempname
	}

	baseEnum, _, _ := using.data().isEnum()

	if assignment == "" {
		code += match.code()
	} else {
		code += assignment
	}

	if !baseEnum.simple {
		code += "->type"
	}

	if caseOf.is == "negate-match-enum" {
		code += " != "
	} else {
		code += " == "
	}

	if caseOf.is == "match-enum" || caseOf.is == "negate-match-enum" {
		matchBaseEnum, matchBaseUn, _ := caseOf.data().isEnum()
		matchBaseEnum = matchBaseEnum.baseEnum()
		enNameSpace := matchBaseEnum.cname
		code += enumTypeName(enNameSpace, matchBaseUn.name)
	} else {
		compare := me.eval(caseOf)
		if compare.data() == nil {
			panic("expected enum but was " + caseOf.string(me.hmfile, 0))
		}
		compareEnum, _, ok := compare.data().isEnum()
		if !ok {
			panic("expected enum but was " + caseOf.string(me.hmfile, 0))
		}
		code += compare.code()
		if !compareEnum.simple {
			code += "->type"
		}
	}

	cb.current = codeNode(n, code)
	return cb
}

func (me *cfile) compileMatch(n *node) *codeblock {
	top := ""
	using := n.has[0]
	match := me.eval(using)

	if match.data().isSomeOrNone() {
		return me.compileMatchSomrOrNone(match, n, top)
	}

	test := match.code()
	tempname := ""

	var isEnum *enum

	if baseEnum, _, ok := using.data().isEnum(); ok {
		isEnum = baseEnum
	}

	if using.is == "variable" {
		name := me.getvar(using.idata.name).cname
		if isEnum != nil && !isEnum.simple {
			test = name
		}
		tempname = name
	}

	cb := &codeblock{}

	ix := 1
	size := len(n.has)
	hasdefault := false

	code := ""
	renaming := ""

	for ix < size {
		caseOf := n.has[ix]
		thenDo := n.has[ix+1]
		if caseOf.is == "_" {
			hasdefault = true
			code += fmc(me.depth) + "default: {\n"
		} else {
			if isEnum != nil {
				if len(caseOf.has) > 0 {
					temphas := caseOf.has[0]
					idata := temphas.idata.name
					if tempname == "" {
						tempname = "match_" + me.temp()
						tempv := temphas.data().getnamedvariable(tempname, false)
						me.scope.variables[tempname] = tempv
						pre := fmtassignspace(match.data().typeSig(me)) + tempname + " = " + test
						cb.prepend(codeNodeUpgrade(hollowCode(pre)))
						test = tempname
					}
					me.scope.renaming[idata] = tempname
					renaming = idata
				}
				code += fmc(me.depth) + "case " + enumTypeName(isEnum.baseEnum().cname, caseOf.is) + ": {\n"

			} else if _, ok := literals[caseOf.is]; ok {
				for hi, h := range caseOf.has {
					code += fmc(me.depth) + "case " + h.is
					if hi == len(caseOf.has)-1 {
						code += ": {\n"
					} else {
						code += ":\n"
					}
				}
			} else {
				code += fmc(me.depth) + "case " + caseOf.is + ": {\n"
			}
		}
		thenBlock := me.eval(thenDo).code()
		if renaming != "" {
			delete(me.scope.renaming, renaming)
		}
		me.depth++
		if thenBlock != "" {
			code += me.maybeFmc(thenBlock, me.depth) + thenBlock + me.maybeColon(thenBlock)
			code += me.maybeNewLine(code)
		}
		if !strings.Contains(thenBlock, "return") {
			code += fmc(me.depth) + "break;\n"
		}
		me.depth--
		code += fmc(me.depth) + "}\n"
		ix += 2
	}
	if !hasdefault {
		code += fmc(me.depth) + "default: exit(1);\n"
	}
	code += fmc(me.depth) + "}"
	if isEnum != nil && !isEnum.simple {
		test += "->type"
	}
	top += "switch (" + test + ") {\n"
	cb.current = codeNode(n, top+code)
	return cb
}

func (me *cfile) compileMatchSomrOrNone(match *codeblock, n *node, code string) *codeblock {
	ifNull := ""
	ifNotNull := ""
	ix := 1
	size := len(n.has)

	using := n.has[0]
	tempname := ""
	casename := ""
	if using.is == "variable" {
		name := me.getvar(using.idata.name).cname
		tempname = name
	}

	for ix < size {
		block := ""
		caseOf := n.has[ix]
		if caseOf.is == "some" {
			if len(caseOf.has) > 0 {
				temphas := caseOf.has[0]
				idata := temphas.idata.name
				if tempname == "" {
					tempname = "match_" + me.temp()
					casename = tempname
					tempv := temphas.data().getnamedvariable(casename, false)
					me.scope.variables[casename] = tempv
				}
				me.scope.renaming[idata] = tempname
			}
		}
		c := me.eval(n.has[ix+1]).code()
		block += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		if casename != "" {
			delete(me.scope.variables, casename)
		}
		if caseOf.is == "some" {
			ifNotNull = block
		} else if caseOf.is == "none" {
			ifNull = block
		}
		ix += 2
	}

	matchcode := match.code()
	boolcode := ""

	if casename != "" {
		code += fmtassignspace(match.data().typeSig(me)) + tempname + " = " + matchcode + ";\n" + fmc(me.depth)
		boolcode = tempname
	} else {
		boolcode = matchcode
	}

	if ifNull != "" && ifNotNull != "" {
		code += "if (" + boolcode + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += me.maybeNewLine(code) + fmc(me.depth) + "} else {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += me.maybeNewLine(code) + fmc(me.depth) + "}"

	} else if ifNull != "" {
		code += "if (" + boolcode + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += me.maybeNewLine(code) + fmc(me.depth) + "}"

	} else {
		code += "if (" + boolcode + " != NULL) {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	}

	return codeBlockOne(n, code)
}
