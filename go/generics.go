package main

func (me *parser) withGenericsHeader() ([]*datatype, map[string][]*classInterface, *parseError) {
	if me.token.is != "with" {
		if me.token.is == "line" && me.peek().is == "with" {
			me.eat("line")
		} else {
			return nil, nil, nil
		}
	}
	me.eat("with")
	module := me.hmfile
	var list []*datatype
	var requirements map[string][]*classInterface
	for {
		gname := me.token.value
		me.wordOrPrimitive()
		data, er := getdatatype(module, gname)
		if er != nil {
			return nil, nil, er
		}
		if me.token.is == ":" {
			me.eat(":")
			interfaces := make([]*classInterface, 0)
			for {
				requires := me.token.value
				me.eat("id")

				moduleReq := module

				if m, ok := module.imports[requires]; ok && me.token.is == "." {
					moduleReq = m
					me.eat(".")
					requires = me.token.value
					me.eat("id")
				}

				interfaceDef, ok := moduleReq.interfaces[requires]
				if !ok {
					return nil, nil, err(me, ECodeMissingInterface, "Missing interface '"+requires+"'")
				}
				for fname := range interfaceDef.functions {
					if def, _, ok := searchInterface(interfaces, fname); ok {
						return nil, nil, err(me, ECodeInterfaceNameConflict, "Conflicting '"+fname+"' from '"+interfaceDef.name+"' and '"+def.name+"'")
					}
				}
				interfaces = append(interfaces, interfaceDef)
				if me.token.is != "," {
					break
				}
				me.eat("+")
			}
			if requirements == nil {
				requirements = make(map[string][]*classInterface)
			}
			requirements[gname] = interfaces
		}
		if list == nil {
			list = make([]*datatype, 0)
		}
		list = append(list, data)
		if me.token.is == "and" {
			me.eat("and")
			continue
		} else if me.token.is == "line" {
			if me.peek().is == "and" {
				me.eat("line")
				me.eat("and")
				continue
			}
			break
		} else {
			return nil, nil, err(me, ECodeUnexpectedToken, "Bad token \""+me.token.is+"\" in class generic.")
		}
	}
	return list, requirements, nil
}

func (me *parser) genericHeader() ([]*datatype, map[string][]*classInterface, *parseError) {
	module := me.hmfile
	var list []*datatype
	var requirements map[string][]*classInterface
	if me.token.is == "<" {
		me.eat("<")
		for {
			gname := me.token.value
			me.wordOrPrimitive()
			data, er := getdatatype(module, gname)
			if er != nil {
				return nil, nil, er
			}
			if me.token.is == ":" {
				me.eat(":")
				interfaces := make([]*classInterface, 0)
				for {
					requires := me.token.value
					me.eat("id")

					moduleReq := module

					if m, ok := module.imports[requires]; ok && me.token.is == "." {
						moduleReq = m
						me.eat(".")
						requires = me.token.value
						me.eat("id")
					}

					interfaceDef, ok := moduleReq.interfaces[requires]
					if !ok {
						return nil, nil, err(me, ECodeMissingInterface, "Missing interface '"+requires+"'")
					}
					for fname := range interfaceDef.functions {
						if def, _, ok := searchInterface(interfaces, fname); ok {
							return nil, nil, err(me, ECodeInterfaceNameConflict, "Conflicting '"+fname+"' from '"+interfaceDef.name+"' and '"+def.name+"'")
						}
					}
					interfaces = append(interfaces, interfaceDef)
					if me.token.is != "+" {
						break
					}
					me.eat("+")
				}
				if requirements == nil {
					requirements = make(map[string][]*classInterface)
				}
				requirements[gname] = interfaces
			}
			if list == nil {
				list = make([]*datatype, 0)
			}
			list = append(list, data)
			if me.token.is == "," {
				me.eat(",")
				continue
			} else if me.token.is == ">" {
				break
			} else {
				return nil, nil, err(me, ECodeUnexpectedToken, "Bad token \""+me.token.is+"\" in class generic.")
			}
		}
		me.eat(">")
	}
	return list, requirements, nil
}

func (me *parser) mapUnionGenerics(en *enum, dict map[string]string) ([]*datatype, *parseError) {
	mapped := make([]*datatype, len(en.generics))
	for i, e := range en.generics {
		to, ok := dict[e]
		if !ok {
			return nil, err(me, ECodeGenericNotImplemented, "Generic \""+e+"\" not implemented for \""+en.name+"\".")
		}
		var er *parseError
		mapped[i], er = getdatatype(me.hmfile, to)
		if er != nil {
			return nil, er
		}
	}
	return mapped, nil
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

func (me *parser) genericsReplacer(module *hmfile, original *datatype, gmapper map[string]string) (*datatype, *parseError) {
	var er *parseError
	data := original.copy()
	if data.generics != nil {
		for i, g := range data.generics {
			data.generics[i], er = me.genericsReplacer(module, g, gmapper)
			if er != nil {
				return nil, er
			}
		}
	}
	if data.parameters != nil {
		for i, p := range data.parameters {
			data.parameters[i], er = me.genericsReplacer(module, p, gmapper)
			if er != nil {
				return nil, er
			}
		}
	}
	if data.variadic != nil {
		data.variadic, er = me.genericsReplacer(module, data.variadic, gmapper)
		if er != nil {
			return nil, er
		}
	}
	if data.returns != nil {
		data.returns, er = me.genericsReplacer(module, data.returns, gmapper)
		if er != nil {
			return nil, er
		}
	}
	if data.member != nil {
		data.member, er = me.genericsReplacer(module, data.member, gmapper)
		if er != nil {
			return nil, er
		}
	}
	data.canonical = mapGenericSingle(data.canonical, gmapper)
	if data.generics != nil {
		implementation := data.print()
		if data.class != nil {
			if cl, ok := data.module.classes[implementation]; ok {
				data.class = cl
			} else {
				data.class, er = me.defineClassImplGeneric(data.class, data.generics)
				if er != nil {
					return nil, er
				}
			}
		} else if data.enum != nil {
			if en, ok := data.module.enums[implementation]; ok {
				data.enum = en
			} else {
				data.enum, er = me.defineEnumImplGeneric(data.enum, data.generics)
				if er != nil {
					return nil, er
				}
			}
		}
	}
	return getdatatype(module, data.print())
}

func hintRecursiveReplace(a, b *datatype, generics []string, update map[string]*datatype) bool {
	if b.is == dataTypeUnknown {
		if inList(generics, b.canonical) >= 0 {
			update[b.canonical] = a
			return true
		}
	}
	if b.is == dataTypeMaybe {
		return hintRecursiveReplace(a, b.member, generics, update)
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
					ok := hintRecursiveReplace(ga, gb, generics, update)
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
			return hintRecursiveReplace(a.member, b, generics, update)
		}
	case dataTypeSlice:
		{
			if b.is != dataTypeSlice {
				return false
			}
			ok := hintRecursiveReplace(a.member, b.member, generics, update)
			if !ok {
				return false
			}
		}
	case dataTypeArray:
		{
			if b.is != dataTypeArray {
				return false
			}
			ok := hintRecursiveReplace(a.member, b.member, generics, update)
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
			ok := hintRecursiveReplace(a.returns, b.returns, generics, update)
			if !ok {
				return false
			}
			for i, pa := range a.parameters {
				pb := b.parameters[i]
				ok := hintRecursiveReplace(pa, pb, generics, update)
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

func (me *parser) hintGeneric(data *datatype, gdata *datatype, generics []string) map[string]*datatype {
	update := make(map[string]*datatype)
	ok := hintRecursiveReplace(data, gdata, generics, update)
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

func inList(ls []string, name string) int {
	for i, s := range ls {
		if s == name {
			return i
		}
	}
	return -1
}
