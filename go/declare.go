package main

import (
	"fmt"
	"strings"
)

func (me *parser) buildClassImplGeneric(typed string, gmapper map[string]string) string {
	fmt.Println("^ build impl generic: \""+typed+"\" with gmapper =>", gmapper)

	base := typed[0:strings.Index(typed, "<")]
	baseClass, ok := me.hmfile.classes[base]
	if !ok {
		panic(me.fail() + "class \"" + base + "\" does not exist")
	}

	order := me.mapGenerics(typed, gmapper)
	impl := base + "<" + strings.Join(order, ",") + ">"
	fmt.Println("$ build impl generic: impl := \"" + impl + "\"")
	if _, ok := me.hmfile.classes[impl]; !ok {
		me.defineClassImplGeneric(baseClass, impl, order)
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

				if _, ok := me.hmfile.classes[pop]; !ok {
					base := me.hmfile.classes[current.name]
					me.defineClassImplGeneric(base, pop, current.order)
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
			return "[]" + me.buildClassImplGeneric(typeOfMem, gmapper)
		}
		return "[]" + me.mapAnyImpl(typeOfMem, gmapper)
	} else if checkHasGeneric(typed) {
		return me.buildClassImplGeneric(typed, gmapper)
	}
	return me.mapAnyImpl(typed, gmapper)
}

func (me *parser) defineEnumImplGeneric(base *enum, impl string, order []string) {
	fmt.Println("define enum impl generic: base \"" + base.name + "\" with impl \"" + impl + "\" and order \"" + strings.Join(order, "|") + "\"")
}

func (me *parser) defineClassImplGeneric(base *class, impl string, order []string) {
	fmt.Println("define class impl generic: base \"" + base.name + "\" with impl \"" + impl + "\" and order \"" + strings.Join(order, "|") + "\"")

	memberMap := make(map[string]*variable)
	for k, v := range base.variables {
		memberMap[k] = v.copy()
	}

	me.hmfile.namespace[impl] = "class"
	me.hmfile.types[impl] = true
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, impl+"_class")

	classDef := classInit(impl, nil, nil)
	classDef.initMembers(base.variableOrder, memberMap)
	me.hmfile.classes[impl] = classDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix]
	}

	for _, mem := range memberMap {
		mem.typed = me.genericsReplacer(mem.typed, gmapper)
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
		_, ok := me.hmfile.types[gimpl]
		if !ok {
			if impl {
				panic(me.fail() + "generic implementation type \"" + gimpl + "\" does not exist")
			}
		}
		order = append(order, gimpl)
	}
	me.eat(">")
	return order
}

func (me *parser) declareType(impl bool) string {
	array := false
	if me.token.is == "[" {
		me.eat("[")
		me.eat("]")
		array = true
	}

	typed := me.token.value
	me.eat("id")

	if _, ok := me.hmfile.imports[typed]; ok {
		me.eat(".")
		typed += "."
		typed += me.token.value
		me.eat("id")
	}

	if me.token.is == "<" {
		module, name := me.hmfile.moduleAndName(typed)
		if base, ok := module.classes[name]; ok {
			gtypes := me.declareGeneric(impl, base)
			typed += "<" + strings.Join(gtypes, ",") + ">"
			if impl {
				fmt.Println("declare type: building class \"" + name + "\" with impl \"" + typed + "\"")
				if _, ok := module.classes[typed]; !ok {
					me.defineClassImplGeneric(base, typed, gtypes)
				}
			}
		} else if base, ok := module.enums[name]; ok {
			gtypes := me.declareGeneric(impl, base)
			typed += "<" + strings.Join(gtypes, ",") + ">"
			if impl {
				fmt.Println("declare type: building enum \"" + name + "\" with impl \"" + typed + "\"")
				if _, ok := module.enums[typed]; !ok {
					me.defineEnumImplGeneric(base, typed, gtypes)
				}
			}
		} else {
			panic(me.fail() + "base enum \"" + name + "\" does not exist")
		}
	}

	if array {
		typed = "[]" + typed
	}

	return typed
}

func (me *parser) nameOfClassFunc(classname, funcname string) string {
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
