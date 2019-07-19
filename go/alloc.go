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
		return me.buildClass(module)
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
			if param.typed != unionType {
				if _, gok := gdict[unionType]; gok {
					gimpl[unionType] = param.typed
				} else {
					panic(me.fail() + "enum \"" + enumName + "\" type \"" + unionName + "\" expects \"" + unionType + "\" but parameter was \"" + param.typed + "\"")
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

// TODO deprecated
func (me *parser) buildClass(module *hmfile) string {
	name := me.token.value
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

	if me.hmfile != module {
		typed = module.name + "." + typed
	}
	return typed
}

func (me *parser) allocClass(module *hmfile) *node {
	n := nodeInit("new")
	// TODO deprecated
	n.typed = me.buildClass(module)
	// n.typed = me.declareType(true)
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
	// TODO deprecated
	n.typed = "[]" + me.buildAnyType()
	// n.typed = "[]" + me.declareType
	n.push(size)
	fmt.Println("array node =", n.string(0))

	return n
}
