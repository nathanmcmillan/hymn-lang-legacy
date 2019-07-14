package main

import (
	"fmt"
	"strings"
)

func (me *parser) buildImplGeneric(typed, gname, gimpl string) string {
	fmt.Println("build impl generic \"" + typed + "\" with gname \"" + gname + "\" and gimpl \"" + gimpl + "\"")

	base := typed[0:strings.Index(typed, "<")]
	baseClass, ok := me.hmfile.classes[base]
	if !ok {
		panic(me.fail() + "class \"" + base + "\" does not exist")
	}

	impl := base + "<"
	gtypes := genericsInType(typed)
	order := make([]string, 0)
	for ix, gtype := range gtypes {
		if ix != 0 {
			impl += ","
		}
		if gtype == gname {
			impl += gimpl
			order = append(order, gimpl)
		} else {
			impl += gtype
			order = append(order, gtype)
		}
	}
	impl += ">"

	fmt.Println("impl generic \"" + impl + "\"")

	if _, ok := me.hmfile.classes[impl]; !ok {
		me.defineImplGeneric(baseClass, impl, order)
	}

	return impl
}

func (me *parser) defineImplGeneric(base *class, impl string, order []string) {
	fmt.Println("define impl generic base \"" + base.name + "\" with impl \"" + impl + "\" and order \"" + strings.Join(order, "|") + "\"")

	memberMap := make(map[string]*variable)
	for k, v := range base.variables {
		memberMap[k] = v.copy()
	}

	me.hmfile.defineOrder = append(me.hmfile.defineOrder, impl+"_class")
	me.hmfile.classes[impl] = classInit(impl, base.variableOrder, memberMap, nil)
	me.hmfile.namespace[impl] = "class"
	me.hmfile.types[impl] = true

	for ix, gname := range base.generics {
		gimpl := order[ix]
		for _, mem := range memberMap {
			if checkIsArray(mem.typed) {
				if typeOfArray(mem.typed) == gname {
					mem.typed = "[]" + gimpl
				} else if checkHasGeneric(mem.typed) {
					mem.typed = "[]" + me.buildImplGeneric(mem.typed, gname, gimpl)
				}
			} else if checkHasGeneric(mem.typed) {
				mem.typed = me.buildImplGeneric(mem.typed, gname, gimpl)
			} else if mem.typed == gname {
				mem.typed = gimpl
			}
		}
	}
}

func (me *parser) eatImplGeneric(gsize int) []string {
	me.eat("<")
	order := make([]string, 0)
	for i := 0; i < gsize; i++ {
		if i != 0 {
			me.eat("delim")
		}
		gimpl := me.declareType()
		if _, ok := me.hmfile.types[gimpl]; !ok {
			panic(me.fail() + "generic implementation type \"" + gimpl + "\" does not exist")
		}
		order = append(order, gimpl)
	}
	me.eat(">")
	return order
}

func (me *parser) declareType() string {
	typed := ""
	if me.token.is == "[" {
		me.eat("[")
		me.eat("]")
		typed += "[]"
	}

	value := me.token.value
	me.eat("id")
	typed += value

	if _, ok := me.hmfile.imports[value]; ok {
		me.eat(".")
		typed += "."
		value = me.token.value
		me.eat("id")
		typed += value
	}

	if me.token.is == "<" {

		module, name := me.hmfile.moduleAndName(typed)
		base, ok := module.classes[name]
		if !ok {
			panic(me.fail() + "base class \"" + name + "\" does not exist")
		}
		gsize := len(base.generics)
		gtypes := me.eatImplGeneric(gsize)

		typed += "<" + strings.Join(gtypes, ",") + ">"

		fmt.Println("declare type: building class \"" + name + "\" with impl \"" + typed + "\"")
		if _, ok := module.classes[typed]; !ok {
			me.defineImplGeneric(base, typed, gtypes)
		}

		// me.eat("<")
		// typed += "<"
		// ix := 0
		// for {
		// 	if ix > 0 {
		// 		typed += "," + me.token.value
		// 	} else {
		// 		typed += me.token.value
		// 	}
		// 	me.eat("id")
		// 	if me.token.is == "delim" {
		// 		me.eat("delim")
		// 		ix++
		// 		continue
		// 	}
		// 	if me.token.is == ">" {
		// 		break
		// 	}
		// 	panic(me.fail() + "bad token \"" + me.token.is + "\" in generic type declaration")
		// }
		// me.eat(">")
		// typed += ">"
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

func genericsInType(typed string) []string {
	parts := strings.Split(typed, "<")
	if len(parts) == 1 {
		return nil
	}
	impl := parts[1]
	impl = impl[0 : len(impl)-1]
	get := strings.Split(impl, ",")
	ls := make([]string, 0)
	for ix := range get {
		ls = append(ls, strings.Trim(get[ix], " "))
	}
	return ls
}

func (me *parser) assignable(n *node) bool {
	return n.is == "variable" || n.is == "member-variable" || n.is == "array-member"
}
