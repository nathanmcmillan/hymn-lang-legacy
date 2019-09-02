package main

import (
	"fmt"
	"strings"
)

func (me *parser) buildAnyType(alloc *allocData) *varData {

	typed := me.token.value
	me.verify("id")

	var module *hmfile
	if _, ok := me.hmfile.imports[typed]; ok {
		module = me.hmfile.program.hmfiles[typed]
		me.eat("id")
		me.eat(".")
		typed = me.token.value
		me.verify("id")
	} else {
		module = me.hmfile
	}

	if _, ok := module.classes[typed]; ok {
		return me.buildClass(nil, module, alloc)
	}

	if _, ok := module.types[typed]; !ok {
		panic(me.fail() + "type \"" + typed + "\" for module \"" + module.name + "\" not found")
	}

	me.eat("id")
	if me.hmfile != module {
		typed = module.name + "." + typed
	}

	vdata := me.hmfile.typeToVarData(typed)
	vdata.merge(alloc)

	return vdata
}

func (me *parser) allocEnum(module *hmfile, alloc *allocData) *node {
	enumName := me.token.value
	me.eat("id")
	enumDef, ok := module.enums[enumName]
	if !ok {
		panic(me.fail() + "enum \"" + enumName + "\" does not exist")
	}

	gdict := enumDef.genericsDict
	var order []string
	if me.token.is == "<" {
		order, _ = me.genericHeader()
		enumName += "<" + strings.Join(order, ",") + ">"
		if len(order) != len(gdict) {
			panic(me.fail() + "generic enum \"" + enumName + "\" with impl " + fmt.Sprint(order) + " does not match " + fmt.Sprint(gdict))
		}
		if _, ok := module.enums[enumName]; !ok {
			me.defineEnumImplGeneric(enumDef, enumName, order)
		}
	}

	me.eat(".")
	unionName := me.token.value
	me.eat("id")
	unionDef, ok := enumDef.types[unionName]
	if !ok {
		panic(me.fail() + "enum \"" + enumName + "\" does not have type \"" + unionName + "\"")
	}

	n := nodeInit("enum")

	typeSize := len(unionDef.types)
	if typeSize > 0 {
		me.eat("(")
		gimpl := make(map[string]string)
		for ix, unionType := range unionDef.types {
			if ix != 0 {
				me.eat("delim")
			}
			param := me.calc(0)
			if param.asVar().notEqual(unionType) {
				if _, gok := gdict[unionType.full]; gok {
					gimpl[unionType.full] = param.getType()
				} else {
					panic(me.fail() + "enum \"" + enumName + "\" type \"" + unionName + "\" expects \"" + unionType.full + "\" but parameter was \"" + param.getType() + "\"")
				}
			}
			n.push(param)
		}
		me.eat(")")
		if len(order) == 0 {
			if len(gimpl) != len(gdict) {
				panic(me.fail() + "generic enum \"" + enumName + "\" with impl " + fmt.Sprint(gimpl) + " does not match " + fmt.Sprint(gdict))
			}
			if len(gimpl) > 0 {
				order = me.mapUnionGenerics(enumDef, gimpl)
				enumName += "<" + strings.Join(order, ",") + ">"
				if _, ok := module.enums[enumName]; !ok {
					me.defineEnumImplGeneric(enumDef, enumName, order)
				}
			}
		}
	} else {
		if len(gdict) != 0 && len(order) == 0 {
			panic(me.fail() + "generic enum \"" + enumName + "\" has no impl for " + fmt.Sprint(enumDef.generics))
		}
	}

	if me.hmfile == module {
		n.vdata = me.hmfile.typeToVarData(enumName)
		n.value = unionName
	} else {
		// n.vdata = module.name + "." + enumName
		n.vdata = module.typeToVarData(enumName)
		n.value = unionName
	}
	return n
}

func defaultValue(typed string) string {
	switch typed {
	case "string":
		return ""
	case "int":
		return "0"
	case "float":
		return "0"
	case "bool":
		return "false"
	default:
		return ""
	}
}

func (me *parser) pushClassParams(n *node, base *class, params []*node) {
	for i, param := range params {
		if param == nil {
			clsvar := base.variables[base.variableOrder[i]]
			dfault := nodeInit(clsvar.vdat.full)
			dfault.copyTypeFromVar(clsvar)
			dfault.value = defaultValue(clsvar.vdat.full)
			n.push(dfault)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) classParams(n *node, typed string) {
	me.eat("(")
	base := me.hmfile.classes[typed]
	vars := base.variableOrder
	params := make([]*node, len(vars))
	pix := 0
	dict := false
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		}
		if pix > 0 || dict {
			me.eat("delim")
		}
		if me.token.is == "id" && me.peek().is == ":" {
			vname := me.token.value
			me.eat("id")
			me.eat(":")
			param := me.calc(0)
			clsvar := base.variables[vname]
			if param.asVar().notEqual(clsvar.vdat) && clsvar.vdat.full != "?" {
				err := "parameter \"" + param.getType()
				err += "\" does not match class \"" + base.name + "\" variable \""
				err += clsvar.name + "\" with type \"" + clsvar.vdat.full + "\""
				panic(me.fail() + err)
			}
			for i, v := range vars {
				if vname == v {
					params[i] = param
					break
				}
			}
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			param := me.calc(0)
			clsvar := base.variables[vars[pix]]
			if param.asVar().notEqual(clsvar.vdat) && clsvar.vdat.full != "?" {
				err := "parameter \"" + param.getType()
				err += "\" does not match class \"" + base.name + "\" variable \""
				err += clsvar.name + "\" with type \"" + clsvar.vdat.full + "\""
				panic(me.fail() + err)
			}
			params[pix] = param
			pix++
		}
	}
	me.pushClassParams(n, base, params)
}

func (me *parser) specialClassParams(depth int, n *node, typed string) {
	ndepth := me.peek().depth
	if ndepth != depth+1 {
		return
	}
	me.eat("line")
	base := me.hmfile.classes[typed]
	vars := base.variableOrder
	params := make([]*node, len(vars))
	for {
		if me.token.is == "id" && me.peek().is == ":" {
			vname := me.token.value
			me.eat("id")
			me.eat(":")
			param := me.calc(0)
			clsvar := base.variables[vname]
			if param.asVar().notEqual(clsvar.vdat) && clsvar.vdat.full != "?" {
				err := "parameter type \"" + param.getType()
				err += "\" does not match class \"" + base.name + "\" with member \""
				err += clsvar.name + "\" and type \"" + clsvar.vdat.full + "\""
				panic(me.fail() + err)
			}
			for i, v := range vars {
				if vname == v {
					params[i] = param
					break
				}
			}
			if me.token.is == "line" {
				ndepth := me.peek().depth
				if ndepth != depth+1 {
					break
				}
				me.eat("line")
				continue
			}
		}
		panic(me.fail() + "missing parameter")
	}
	me.pushClassParams(n, base, params)
}

func (me *parser) buildClass(n *node, module *hmfile, alloc *allocData) *varData {
	name := me.token.value
	depth := me.token.depth
	me.eat("id")
	base, ok := module.classes[name]
	if !ok {
		panic(me.fail() + "class \"" + name + "\" does not exist")
	}
	typed := name
	gsize := len(base.generics)
	if gsize > 0 && me.token.is == "<" {
		gtypes := me.declareGeneric(true, base)
		typed = name + "<" + strings.Join(gtypes, ",") + ">"
		fmt.Println("building class \"" + name + "\" with impl \"" + typed + "\"")
		if _, ok := me.hmfile.classes[typed]; !ok {
			me.defineClassImplGeneric(base, typed, gtypes)
		}
	}
	if n != nil {
		if me.token.is == "(" {
			me.classParams(n, typed)
		} else if me.token.is == "line" {
			me.specialClassParams(depth, n, typed)
		}
	}
	if me.hmfile != module {
		typed = module.name + "." + typed
	}

	vdata := me.hmfile.typeToVarData(typed)
	vdata.merge(alloc)

	return vdata
}

func (me *parser) allocClass(module *hmfile, alloc *allocData) *node {
	n := nodeInit("new")
	n.vdata = me.buildClass(n, module, alloc)
	if alloc != nil && alloc.useStack {
		n.attributes["use-stack"] = "true"
	}
	return n
}
