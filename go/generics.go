package main

func (me *parser) genericHeader() ([]*datatype, map[string]int) {
	order := make([]*datatype, 0)
	dict := make(map[string]int)
	if me.token.is == "<" {
		me.eat("<")
		for {
			gname := me.token.value
			me.wordOrPrimitive()
			dict[gname] = len(order)
			order = append(order, getdatatype(me.hmfile, gname))
			if me.token.is == "," {
				me.eat(",")
				continue
			}
			if me.token.is == ">" {
				break
			}
			panic(me.fail() + "Bad token \"" + me.token.is + "\" in class generic.")
		}
		me.eat(">")
	}
	return order, dict
}

func (me *parser) mapUnionGenerics(en *enum, dict map[string]string) []*datatype {
	mapped := make([]*datatype, len(en.generics))
	for i, e := range en.generics {
		to, ok := dict[e]
		if !ok {
			panic(me.fail() + "Generic \"" + e + "\" not implemented for \"" + en.name + "\".")
		}
		mapped[i] = getdatatype(me.hmfile, to)
	}
	return mapped
}

type gstack struct {
	name  string
	order []string
}

func mapGenericSingle(typed string, gmapper map[string]string) string {
	implementation, ok := gmapper[typed]
	if ok {
		return implementation
	}
	return typed
}

func (me *parser) genericsReplacer(module *hmfile, original *datatype, gmapper map[string]string) *datatype {
	data := original.copy()
	if data.generics != nil {
		for i, g := range data.generics {
			data.generics[i] = me.genericsReplacer(module, g, gmapper)
		}
	}
	if data.parameters != nil {
		for i, p := range data.parameters {
			data.parameters[i] = me.genericsReplacer(module, p, gmapper)
		}
	}
	if data.variadic != nil {
		data.variadic = me.genericsReplacer(module, data.variadic, gmapper)
	}
	if data.returns != nil {
		data.returns = me.genericsReplacer(module, data.returns, gmapper)
	}
	if data.member != nil {
		data.member = me.genericsReplacer(module, data.member, gmapper)
	}
	data.canonical = mapGenericSingle(data.canonical, gmapper)
	if data.generics != nil {
		implementation := data.print()
		if data.class != nil {
			if cl, ok := data.module.classes[implementation]; ok {
				data.class = cl
			} else {
				data.class = me.defineClassImplGeneric(data.class, data.generics)
			}
		} else if data.enum != nil {
			if en, ok := data.module.enums[implementation]; ok {
				data.enum = en
			} else {
				data.enum = me.defineEnumImplGeneric(data.enum, data.generics)
			}
		}
	}
	return getdatatype(module, data.print())
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

func (me *parser) hintGeneric(data *datatype, gdata *datatype, gindex map[string]int) map[string]*datatype {
	update := make(map[string]*datatype)
	ok := hintRecursiveReplace(data, gdata, gindex, update)
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
