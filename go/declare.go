package main

import (
	"fmt"
	"strings"
)

func (me *parser) mapUnionGenerics(en *enum, dict map[string]string) []string {
	mapped := make([]string, len(en.generics))
	for i, e := range en.generics {
		to, ok := dict[e]
		if !ok {
			panic(me.fail() + "generic \"" + e + "\" not implemented for \"" + en.name + "\"")
		}
		mapped[i] = to
	}
	return mapped
}

func (me *parser) buildImplGeneric(typed string, gmapper map[string]string) string {
	fmt.Println("^ build impl generic: \""+typed+"\" with gmapper =>", gmapper)

	base := typed[0:strings.Index(typed, "<")]

	var baseEnum *enum
	baseClass, okc := me.hmfile.classes[base]

	if !okc {
		var oke bool
		baseEnum, oke = me.hmfile.enums[base]
		if !oke {
			panic(me.fail() + "type \"" + base + "\" does not exist")
		}
	}

	order := me.mapGenerics(typed, gmapper)
	impl := base + "<" + strings.Join(order, ",") + ">"
	fmt.Println("$ build impl generic: impl := \"" + impl + "\"")

	if okc {
		if _, ok := me.hmfile.classes[impl]; !ok {
			me.defineClassImplGeneric(baseClass, impl, order)
		}
	} else if _, ok := me.hmfile.enums[impl]; !ok {
		me.defineEnumImplGeneric(baseEnum, impl, order)
	}

	return impl
}

type gstack struct {
	name  string
	order []string
}

func (me *parser) mapGenerics(typed string, gmapper map[string]string) []string {

	var order []string
	stack := make([]*gstack, 0)
	rest := typed

	for {
		begin := strings.Index(rest, "<")
		end := strings.Index(rest, ">")
		comma := strings.Index(rest, ",")

		if begin != -1 && (begin < end || end == -1) && (begin < comma || comma == -1) {
			name := rest[:begin]
			current := &gstack{}
			current.name = name
			stack = append(stack, current)
			rest = rest[begin+1:]

		} else if end != -1 && (end < begin || begin == -1) && (end < comma || comma == -1) {
			size := len(stack) - 1
			current := stack[size]
			if end == 0 {
			} else {
				sub := rest[:end]
				current.order = append(current.order, me.mapAnyImpl(sub, gmapper))
			}
			stack = stack[:size]
			if size == 0 {
				order = current.order
				break
			} else {
				pop := current.name + "<" + strings.Join(current.order, ",") + ">"

				if _, okc := me.hmfile.classes[pop]; !okc {
					if _, oke := me.hmfile.enums[pop]; oke {
						base := me.hmfile.enums[current.name]
						me.defineEnumImplGeneric(base, pop, current.order)
					} else {
						base := me.hmfile.classes[current.name]
						me.defineClassImplGeneric(base, pop, current.order)
					}
				}

				next := stack[len(stack)-1]
				next.order = append(next.order, pop)
			}
			if end == 0 {
				rest = rest[1:]
			} else {
				rest = rest[end+1:]
			}

		} else if comma != -1 && (comma < begin || begin == -1) && (comma < end || end == -1) {
			current := stack[len(stack)-1]
			if comma == 0 {
				rest = rest[1:]
				continue
			}
			sub := rest[:comma]
			current.order = append(current.order, me.mapAnyImpl(sub, gmapper))
			rest = rest[comma+1:]

		} else {
			panic(me.fail() + "could not parse impl of type \"" + typed + "\"")
		}
	}

	fmt.Println("map generics: \"" + strings.Join(order, "|") + "\"")
	return order
}

func (me *parser) mapAnyImpl(mem string, gmapper map[string]string) string {
	impl, ok := gmapper[mem]
	if ok {
		return impl
	}
	return mem
}

func (me *parser) genericsReplacer(typed string, gmapper map[string]string) string {
	fmt.Println("replacer: \""+typed+"\" =>", gmapper)
	if checkIsArray(typed) {
		typeOfMem := typeOfArray(typed)
		if checkHasGeneric(typed) {
			return "[]" + me.buildImplGeneric(typeOfMem, gmapper)
		}
		return "[]" + me.mapAnyImpl(typeOfMem, gmapper)
	} else if checkHasGeneric(typed) {
		return me.buildImplGeneric(typed, gmapper)
	}
	return me.mapAnyImpl(typed, gmapper)
}

func (me *parser) defineEnumImplGeneric(base *enum, impl string, order []string) {
	fmt.Println("define enum impl generic: base \"" + base.name + "\" with impl \"" + impl + "\" and order \"" + strings.Join(order, "|") + "\"")

	unionList := make([]*union, len(base.types))
	unionDict := make(map[string]*union)
	for i, v := range base.typesOrder {
		cp := v.copy()
		unionList[i] = cp
		unionDict[cp.name] = cp
	}

	me.hmfile.namespace[impl] = "enum"
	me.hmfile.types[impl] = true
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, impl+"_enum")

	enumDef := enumInit(base.module, impl, false, unionList, unionDict, nil, nil)
	me.hmfile.enums[impl] = enumDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix]
	}

	for _, un := range unionList {
		for i, typed := range un.types {
			un.types[i] = me.hmfile.typeToVarData(me.genericsReplacer(typed.full, gmapper))
		}
	}
}

func (me *parser) defineClassImplGeneric(base *class, impl string, order []string) {
	memberMap := make(map[string]*variable)
	for k, v := range base.variables {
		memberMap[k] = v.copy()
	}

	me.hmfile.namespace[impl] = "type"
	me.hmfile.types[impl] = true
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, impl+"_type")

	classDef := classInit(impl, nil, nil)
	classDef.initMembers(base.variableOrder, memberMap)
	me.hmfile.classes[impl] = classDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix]
	}

	for _, mem := range memberMap {
		mem.update(me.hmfile, me.genericsReplacer(mem.vdat.full, gmapper))
	}
}

func (me *parser) declareGeneric(impl bool, base hasGenerics) []string {
	me.eat("<")
	gsize := len(base.getGenerics())
	order := make([]string, 0)
	for i := 0; i < gsize; i++ {
		if i != 0 {
			me.eat("delim")
		}
		gimpl := me.declareType(impl)
		_, ok := me.hmfile.types[gimpl.full]
		if !ok {
			if impl {
				panic(me.fail() + "generic implementation type \"" + gimpl.full + "\" does not exist")
			}
		}
		order = append(order, gimpl.full)
	}
	me.eat(">")
	return order
}

func (me *parser) declareFn() *varData {
	fmt.Println("DECLARE FN ::")
	me.eat("(")
	fn := fnSigInit(me.hmfile)
	if me.token.is != ")" {
		for {
			typed := me.declareType(true)
			fn.args = append(fn.args, fnArgInit(typed.asVariable()))
			if me.token.is == ")" {
				break
			} else if me.token.is == "delim" {
				me.eat("delim")
				continue
			}
			panic(me.fail() + "unexpected token in function pointer")
		}
	}
	me.eat(")")
	if me.token.is != "line" && me.token.is != "," {
		fn.typed = me.declareType(true)
	} else {
		fn.typed = me.hmfile.typeToVarData("void")
	}

	return fn.asVar()
}

func (me *parser) declareFnPtr(fn *function) *varData {
	fmt.Println("DECLARE FN PTR ::", fn.name)
	return me.hmfile.typeToVarData(fn.name)
}

func (me *parser) declareType(impl bool) *varData {
	array := false
	if me.token.is == "[" {
		me.eat("[")
		me.eat("]")
		array = true
	}

	typed := ""

	if me.token.is == "(" {
		return me.declareFn()

	} else if me.token.is == "maybe" {
		me.eat("maybe")
		me.eat("<")
		option := me.declareType(impl).typed
		me.eat(">")
		typed += "maybe<" + option + ">"

	} else if me.token.is == "none" {
		me.eat("none")
		me.eat("<")
		option := me.declareType(impl).typed
		me.eat(">")
		typed += "none<" + option + ">"

	} else {
		typed += me.token.value
		me.eat("id")
	}

	if _, ok := me.hmfile.imports[typed]; ok {
		me.eat(".")
		typed += "."
		typed += me.token.value
		me.eat("id")
	}

	if _, ok := me.hmfile.enums[typed]; ok && me.token.is == "." {
		me.eat(".")
		typed += "."
		typed += me.token.value
		me.eat("id")
	}

	if fn, ok := me.hmfile.functions[typed]; ok {
		return me.declareFnPtr(fn)
	}

	if me.token.is == "<" {
		data := me.hmfile.typeToVarData(typed)
		if base, ok := data.module.classes[data.typed]; ok {
			gtypes := me.declareGeneric(impl, base)
			typed += "<" + strings.Join(gtypes, ",") + ">"
			if impl {
				fmt.Println("declare type: building class \"" + data.typed + "\" with impl \"" + typed + "\"")
				if _, ok := data.module.classes[typed]; !ok {
					me.defineClassImplGeneric(base, typed, gtypes)
				}
			}
		} else if base, ok := data.module.enums[data.typed]; ok {
			gtypes := me.declareGeneric(impl, base)
			typed += "<" + strings.Join(gtypes, ",") + ">"
			if impl {
				fmt.Println("declare type: building enum \"" + data.typed + "\" with impl \"" + typed + "\"")
				if _, ok := data.module.enums[typed]; !ok {
					me.defineEnumImplGeneric(base, typed, gtypes)
				}
			}
		} else {
			panic(me.fail() + "type \"" + data.typed + "\" does not exist")
		}
	}

	if array {
		typed = "[]" + typed
	}

	return me.hmfile.typeToVarData(typed)
}

func nameOfClassFunc(classname, funcname string) string {
	return classname + "_" + funcname
}

func typeOfArray(typed string) string {
	return typed[2:]
}

func checkIsArray(typed string) bool {
	return strings.HasPrefix(typed, "[]")
}

func checkHasGeneric(typed string) bool {
	return strings.HasSuffix(typed, ">")
}

func (me *parser) assignable(n *node) bool {
	return n.is == "variable" || n.is == "member-variable" || n.is == "array-member"
}
