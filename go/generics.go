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
	base := typed[0:strings.Index(typed, "<")]

	var baseEnum *enum
	baseClass, okc := me.hmfile.classes[base]
	baseEnum, oke := me.hmfile.enums[base]

	if !okc && !oke && base != "maybe" {
		panic(me.fail() + "type \"" + base + "\" does not exist")
	}

	order := me.mapGenerics(typed, gmapper)
	impl := base + "<" + strings.Join(order, ",") + ">"

	if okc {
		if _, ok := me.hmfile.classes[impl]; !ok {
			me.defineClassImplGeneric(baseClass, impl, order)
		}
	} else if oke {
		if _, ok := me.hmfile.enums[impl]; !ok {
			me.defineEnumImplGeneric(baseEnum, impl, order)
		}
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
				current.order = append(current.order, me.mapGenericSingle(sub, gmapper))
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
			current.order = append(current.order, me.mapGenericSingle(sub, gmapper))
			rest = rest[comma+1:]

		} else {
			panic(me.fail() + "could not parse impl of type \"" + typed + "\"")
		}
	}

	return order
}

func (me *parser) mapGenericSingle(mem string, gmapper map[string]string) string {
	impl, ok := gmapper[mem]
	if ok {
		return impl
	}
	return mem
}

func (me *parser) mapGenericFunctionSig(typed string, gmapper map[string]string) string {
	args, ret := functionSigType(typed)
	sig := "("
	for i, a := range args {
		if i > 0 {
			sig += ", "
		}
		sig += me.mapGenericSingle(a, gmapper)
	}
	sig += ") " + me.mapGenericSingle(ret, gmapper)
	return sig
}

func (me *parser) genericsReplacer(typed string, gmapper map[string]string) string {
	if checkIsArrayOrSlice(typed) {
		size, typeOfMem := typeOfArrayOrSlice(typed)
		if checkHasGeneric(typed) {
			return "[" + size + "]" + me.buildImplGeneric(typeOfMem, gmapper)
		}
		return "[" + size + "]" + me.mapGenericSingle(typeOfMem, gmapper)
	} else if checkHasGeneric(typed) {
		return me.buildImplGeneric(typed, gmapper)
	} else if checkIsFunction(typed) {
		return me.mapGenericFunctionSig(typed, gmapper)
	}
	return me.mapGenericSingle(typed, gmapper)
}

func hintRecursiveReplace(a, b *datatype, gindex map[string]int, update map[string]string) bool {
	switch a.is {
	case dataTypePrimitive:
		{
			if b.is == dataTypePrimitive {
				return true
			}
		}
	case dataTypeMaybe:
		{
			return hintRecursiveReplace(a.member, b, gindex, update)
		}
	case dataTypeArray:
		{
		}
	case dataTypeFunction:
		{
		}
	case dataTypeClass:
		{
		}
	case dataTypeEnum:
		{
		}
	default:
		panic("missing data type")
	}
	return false
}

func (me *parser) hintGeneric(data *varData, gdata *varData, gindex map[string]int) (bool, map[string]string) {

	fmt.Println("HINT GENERIC INGEST ::", data.full, "|", gdata.full, "|", gindex)

	a := me.hmfile.getdatatype(data.full)
	b := me.hmfile.getdatatype(gdata.full)

	fmt.Println("TYPE SIMPLIFY ::", data.full, "->", a.print())
	fmt.Println("TYPE SIMPLIFY ::", gdata.full, "->", b.print())

	update := make(map[string]string)

	ok := hintRecursiveReplace(a, b, gindex, update)

	return ok, update

	// length := len(w)

	// if length != len(u) || len(t) != len(v) {
	// 	return false, nil
	// }

	// update := make(map[string]string)

	// i := 0
	// for i < length {
	// 	if w[i] != u[i] {
	// 		return false, nil
	// 	}
	// 	i++
	// }
	// i = 0
	// length = len(t)
	// for i < length {
	// 	g := v[i]
	// 	if _, ok := gindex[g]; ok {
	// 		if e, exist := update[g]; exist {
	// 			if e != t[i] {
	// 				return false, nil
	// 			}
	// 		}
	// 		update[g] = t[i]
	// 	} else if t[i] != g {
	// 		return false, nil
	// 	}
	// 	i++
	// }

	// parse out full type string
	// generate list ordering of types plus the in between strings
	// combine to generate full type implementation plus list of individual types

	// do this for impl and base
	// generate varData using reconstructed signature
	// compare varData types

	// (v) v
	// (string) string

	// [3]v
	// [3]string

	// maybe<v>
	// string

	// [3]maybe<v>
	// [3]string

	// [3]data.hashmap<animals, maybe<math.bigint>>
}

func mergeMaps(one, two map[string]string) (bool, map[string]string) {
	merge := make(map[string]string)
	for k, v := range one {
		w, exist := two[k]
		if exist {
			if v == w {
				continue
			}
			return false, nil
		}
		merge[k] = v
	}
	for k, v := range two {
		if _, exist := merge[k]; exist {
			merge[k] = v
		}
	}
	return true, merge
}
