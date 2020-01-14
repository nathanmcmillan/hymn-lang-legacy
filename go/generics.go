package main

import (
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

func (me *parser) buildImplGeneric(module *hmfile, typed string, gmapper map[string]string) string {

	base := typed[0:strings.Index(typed, "<")]

	var baseEnum *enum
	baseClass, okc := module.classes[base]
	baseEnum, oke := module.enums[base]

	if !okc && !oke && base != "maybe" {
		panic(me.fail() + "type \"" + base + "\" does not exist")
	}

	order := me.mapGenerics(module, typed, gmapper)
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

func (me *parser) mapGenerics(module *hmfile, typed string, gmapper map[string]string) []string {
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
				current.order = append(current.order, mapGenericSingle(sub, gmapper))
			}
			stack = stack[:size]
			if size == 0 {
				order = current.order
				break
			} else {
				pop := current.name + "<" + strings.Join(current.order, ",") + ">"

				if _, okc := module.classes[pop]; !okc {
					if _, oke := module.enums[pop]; oke {
						base := module.enums[current.name]
						me.defineEnumImplGeneric(base, pop, current.order)
					} else {
						base := module.classes[current.name]
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
			current.order = append(current.order, mapGenericSingle(sub, gmapper))
			rest = rest[comma+1:]

		} else {
			panic(me.fail() + "could not parse impl of type \"" + typed + "\"")
		}
	}

	return order
}

func mapGenericSingle(mem string, gmapper map[string]string) string {
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
		sig += mapGenericSingle(a, gmapper)
	}
	sig += ") " + mapGenericSingle(ret, gmapper)
	return sig
}

func (me *parser) genericsReplacer(vdata *varData, gmapper map[string]string) string {

	correct := true
	if correct {
		typed := vdata.getRaw()
		module := vdata.getmodule()
		if checkIsArrayOrSlice(typed) {
			size, typeOfMem := typeOfArrayOrSlice(typed)
			if checkHasGeneric(typed) {
				return "[" + size + "]" + me.buildImplGeneric(module, typeOfMem, gmapper)
			}
			return "[" + size + "]" + mapGenericSingle(typeOfMem, gmapper)
		} else if checkHasGeneric(typed) {
			return me.buildImplGeneric(module, typed, gmapper)
		} else if checkIsFunction(typed) {
			return me.mapGenericFunctionSig(typed, gmapper)
		}
		return mapGenericSingle(typed, gmapper)
	}

	data := vdata.dtype
	if data.isSome() {
		member := data.member
		return "maybe<" + me.buildImplGeneric(member.module, member.print(), gmapper) + ">"
	} else if data.isNone() {
		member := data.member
		if member == nil {
			return "none"
		}
		return "none<" + me.buildImplGeneric(member.module, member.print(), gmapper) + ">"
	} else if data.isArrayOrSlice() {
		member := data.member
		if member.generics != nil {
			return "[" + data.arraySize() + "]" + me.buildImplGeneric(member.module, member.print(), gmapper)
		}
		return "[" + data.arraySize() + "]" + mapGenericSingle(member.print(), gmapper)
	} else if data.generics != nil {
		return me.buildImplGeneric(data.module, data.print(), gmapper)
	} else if data.isFunction() {
		return me.mapGenericFunctionSig(data.print(), gmapper)
	}
	return mapGenericSingle(data.print(), gmapper)

	// switch me.is {
	// case dataTypeMaybe:
	// 	{
	// 		return "maybe<" + parser.buildImplGeneric(me.module, me.member.print(), gmapper) + ">"
	// 	}
	// case dataTypeNone:
	// 	{
	// 		if me.member != nil {
	// 			return "none<" + parser.buildImplGeneric(me.module, me.member.print(), gmapper) + ">"
	// 		}
	// 		return "none"
	// 	}
	// case dataTypeArray:
	// 	{
	// 		return "[" + me.size + "]" + me.member.print()
	// 	}
	// case dataTypeSlice:
	// 	{
	// 		return "[]" + me.member.print()
	// 	}
	// case dataTypeFunction:
	// 	{
	// 		return parser.mapGenericFunctionSig(me.print(), gmapper)
	// 	}
	// case dataTypeClass:
	// 	{
	// 		return ""
	// 	}
	// case dataTypeEnum:
	// 	{
	// 		return ""
	// 	}
	// default:
	// 	return mapGenericSingle(me.print(), gmapper)
	// }
}

func hintRecursiveReplace(a, b *datatype, gindex map[string]int, update map[string]*datatype) bool {
	if b.is == dataTypeUnknown {
		if _, ok := gindex[b.canonical]; ok {
			update[b.canonical] = a
			return true
		}
	}
	if b.is == dataTypeMaybe {
		return hintRecursiveReplace(a, b.member, gindex, update)
	}
	switch a.is {
	case dataTypeClass:
		fallthrough
	case dataTypeEnum:
		fallthrough
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypePrimitive:
		{
			if a.generics != nil || b.generics != nil {
				if a.generics == nil || b.generics == nil {
					return false
				}
				if len(a.generics) != len(b.generics) {
					return false
				}
				for i, ga := range a.generics {
					gb := b.generics[i]
					ok := hintRecursiveReplace(ga, gb, gindex, update)
					if !ok {
						return false
					}
				}
			}
		}
	case dataTypeNone:
		{
			return b.is == dataTypeNone
		}
	case dataTypeMaybe:
		{
			return hintRecursiveReplace(a.member, b, gindex, update)
		}
	case dataTypeSlice:
		{
			if b.is != dataTypeSlice {
				return false
			}
			ok := hintRecursiveReplace(a.member, b.member, gindex, update)
			if !ok {
				return false
			}
		}
	case dataTypeArray:
		{
			if b.is != dataTypeArray {
				return false
			}
			ok := hintRecursiveReplace(a.member, b.member, gindex, update)
			if !ok {
				return false
			}
		}
	case dataTypeFunction:
		{
			if b.is != dataTypeFunction {
				return false
			}
			if len(a.parameters) != len(b.parameters) {
				return false
			}
			ok := hintRecursiveReplace(a.returns, b.returns, gindex, update)
			if !ok {
				return false
			}
			for i, pa := range a.parameters {
				pb := b.parameters[i]
				ok := hintRecursiveReplace(pa, pb, gindex, update)
				if !ok {
					return false
				}
			}
		}
	default:
		panic("missing data type " + a.nameIs())
	}
	return true
}

func (me *parser) hintGeneric(data *varData, gdata *varData, gindex map[string]int) map[string]*datatype {
	update := make(map[string]*datatype)
	ok := hintRecursiveReplace(data.dtype, gdata.dtype, gindex, update)
	if !ok {
		return nil
	}
	return update
}

func mergeMaps(one, two map[string]*datatype) (bool, map[string]*datatype) {
	merge := make(map[string]*datatype)
	for k, v := range one {
		w, ok := two[k]
		if ok && v.notEquals(w) {
			return false, nil
		}
		merge[k] = v
	}
	for k, v := range two {
		if _, ok := merge[k]; !ok {
			merge[k] = v
		}
	}
	return true, merge
}
