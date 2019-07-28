package main

import (
	"fmt"
	"strings"
)

// TODO deprecated
func (me *parser) buildAnyType() string {

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
		return me.buildClass(nil, module)
	}

	if _, ok := module.types[typed]; !ok {
		panic(me.fail() + "type \"" + typed + "\" for module \"" + module.name + "\" not found")
	}

	me.eat("id")
	if me.hmfile != module {
		typed = module.name + "." + typed
	}

	return typed
}

func (me *parser) allocEnum(module *hmfile) *node {
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
			param := me.calc()
			if me.hmfile.typeToVarData(param.typed).notEqual(unionType) {
				if _, gok := gdict[unionType.full]; gok {
					gimpl[unionType.full] = param.typed
				} else {
					panic(me.fail() + "enum \"" + enumName + "\" type \"" + unionName + "\" expects \"" + unionType.full + "\" but parameter was \"" + param.typed + "\"")
				}
			}
			n.push(param)
		}
		me.eat(")")
		fmt.Println(enumName, unionName, order, gimpl, gdict)
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
		n.typed = enumName
		n.value = unionName
	} else {
		n.typed = module.name + "." + enumName
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
			dfault := nodeInit(clsvar.typed)
			dfault.typed = clsvar.typed
			dfault.value = defaultValue(clsvar.typed)
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
			param := me.calc()
			clsvar := base.variables[vname]
			if me.hmfile.typeToVarData(param.typed).notEqual(clsvar.vdat) && clsvar.typed != "?" {
				err := "parameter \"" + param.typed
				err += "\" does not match class \"" + base.name + "\" variable \""
				err += clsvar.name + "\" with type \"" + clsvar.typed + "\""
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
			param := me.calc()
			clsvar := base.variables[vars[pix]]
			if me.hmfile.typeToVarData(param.typed).notEqual(clsvar.vdat) && clsvar.typed != "?" {
				err := "parameter \"" + param.typed
				err += "\" does not match class \"" + base.name + "\" variable \""
				err += clsvar.name + "\" with type \"" + clsvar.typed + "\""
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
			param := me.calc()
			clsvar := base.variables[vname]
			if me.hmfile.typeToVarData(param.typed).notEqual(clsvar.vdat) && clsvar.typed != "?" {
				err := "parameter type \"" + param.typed
				err += "\" does not match class \"" + base.name + "\" with member \""
				err += clsvar.name + "\" and type \"" + clsvar.typed + "\""
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

func (me *parser) buildClass(n *node, module *hmfile) string {
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
	return typed
}

func (me *parser) allocClass(module *hmfile) *node {
	n := nodeInit("new")
	n.typed = me.buildClass(n, module)
	return n
}

func (me *parser) allocArray() *node {
	me.eat("[")
	size := me.calc()
	if size.typed != "int" {
		panic(me.fail() + "array size must be integer")
	}
	me.eat("]")

	n := nodeInit("array")
	n.typed = "[]" + me.buildAnyType()
	n.push(size)
	fmt.Println("array node =", n.string(0))

	return n
}
