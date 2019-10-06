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
	fmt.Println("REPLACE :: \""+typed+"\" =>", gmapper)
	if checkIsArrayOrSlice(typed) {
		size, typeOfMem := typeOfArrayOrSlice(typed)
		if checkHasGeneric(typed) {
			return "[" + size + "]" + me.buildImplGeneric(typeOfMem, gmapper)
		}
		return "[" + size + "]" + me.mapAnyImpl(typeOfMem, gmapper)
	} else if checkHasGeneric(typed) {
		return me.buildImplGeneric(typed, gmapper)
	}
	return me.mapAnyImpl(typed, gmapper)
}
