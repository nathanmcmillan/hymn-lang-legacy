package main

import (
	"fmt"
	"strings"
)

func (me *parser) defineEnumImplGeneric(base *enum, impl string, order []*datatype) *enum {

	unionList := make([]*union, len(base.types))
	unionDict := make(map[string]*union)
	for i, v := range base.typesOrder {
		cp := v.copy()
		unionList[i] = cp
		unionDict[cp.name] = cp
	}

	me.hmfile.namespace[impl] = "enum"
	me.hmfile.types[impl] = "enum"
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, impl+"_enum")

	enumDef := enumInit(base.module, impl, false, unionList, unionDict, nil, nil)
	enumDef.base = base
	base.impls = append(base.impls, enumDef)
	me.hmfile.enums[impl] = enumDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix].getRaw()
	}
	enumDef.gmapper = gmapper

	for _, un := range unionList {
		for i, data := range un.types {
			un.types[i] = getdatatype(me.hmfile, me.genericsReplacer(data, gmapper).print())
		}
	}

	return enumDef
}

func (me *parser) defineClassImplGeneric(base *class, impl string, order []*datatype) *class {
	memberMap := make(map[string]*variable)
	for k, v := range base.variables {
		memberMap[k] = v.copy()
	}

	module := base.module

	fmt.Println("inserting new class ::", impl)

	module.namespace[impl] = "type"
	module.types[impl] = "class"
	module.defineOrder = append(module.defineOrder, impl+"_type")

	classDef := classInit(module, impl, nil, nil)
	classDef.base = base
	base.impls = append(base.impls, classDef)
	classDef.initMembers(base.variableOrder, memberMap)
	module.classes[impl] = classDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		fmt.Println(impl, "generic map ::", gname, "<-", order[ix].getRaw())
		gmapper[gname] = order[ix].getRaw()
	}
	classDef.gmapper = gmapper

	for _, mem := range memberMap {
		replace := me.genericsReplacer(mem.data(), gmapper).print()
		data := getdatatype(module, replace)
		clname := ""
		cl, ok := data.isClass()
		if ok {
			clname = " | " + cl.name
		}
		fmt.Println(impl, "replacing member ::", mem.name, "<-", data.print()+clname)
		mem._vdata = data
	}

	for _, fn := range base.functionOrder {
		remapClassFunctionImpl(classDef, fn)
	}

	return classDef
}

func (me *parser) declareGeneric(implementation bool, base hasGenerics) []*datatype {
	me.eat("<")
	gsize := len(base.getGenerics())
	order := make([]*datatype, 0)
	for i := 0; i < gsize; i++ {
		if i != 0 {
			me.eat(",")
		}
		gimpl := me.declareType(implementation)
		order = append(order, gimpl)
	}
	me.eat(">")
	return order
}

func (me *parser) declareFn() *datatype {
	me.eat("(")
	fn := fnSigInit(me.hmfile)
	if me.token.is != ")" {
		for {
			typed := me.declareType(true)
			fn.args = append(fn.args, fnArgInit(typed.getvariable()))
			if me.token.is == ")" {
				break
			} else if me.token.is == "," {
				me.eat(",")
				continue
			}
			panic(me.fail() + "unexpected token in function pointer")
		}
	}
	me.eat(")")
	if me.token.is != "line" && me.token.is != "," {
		fn.returns = me.declareType(true)
	} else {
		fn.returns = getdatatype(me.hmfile, "void")
	}

	return fn.newdatatype()
}

func (me *parser) declareFnPtr(fn *function) *datatype {
	return getdatatype(me.hmfile, fn.name)
}

func (me *parser) declareType(implementation bool) *datatype {

	if me.token.is == "[" {
		me.eat("[")
		if me.token.is == "]" {
			me.eat("]")
			return newdataslice(me.declareType(implementation))
		}
		sizeNode := me.calc(0, nil)
		if sizeNode.value == "" || !sizeNode.data().isInt() {
			panic(me.fail() + "array size must be constant integer")
		}
		me.eat("]")
		return newdataarray(sizeNode.value, me.declareType(implementation))
	}

	module := me.hmfile
	value := ""

	if me.token.is == "(" {
		return me.declareFn()

	} else if me.token.is == "maybe" {
		me.eat("maybe")
		me.eat("<")
		option := me.declareType(implementation)
		me.eat(">")
		return newdatamaybe(option)

	} else if me.token.is == "none" {
		me.eat("none")
		if me.token.is == "<" {
			me.eat("<")
			option := me.declareType(implementation)
			me.eat(">")
			return newdatamaybe(option)
		}
		return newdatanone()
	} else {
		value += me.token.value
		me.wordOrPrimitive()
	}

	if value == "void" {
		return newdatavoid()
	}

	if value == "?" {
		return newdataunknown(nil, nil, value, nil)
	}

	if value == TokenString {
		return newdatastring()
	}

	if checkIsPrimitive(value) {
		return newdataprimitive(value)
	}

	original := value

	if strings.HasPrefix(value, "%") {
		if m, ok := me.hmfile.program.modules[value]; ok && me.token.is == "." {
			me.eat(".")
			module = m
			value = module.uidref(me.token.value)
			me.eat("id")
		}
	} else if m, ok := me.hmfile.imports[value]; ok && me.token.is == "." {
		me.eat(".")
		module = m
		value = module.uidref(me.token.value)
		me.eat("id")
	} else {
		value = module.uidref(value)
	}

	fmt.Println("value ::", value)

	if fn, ok := module.functions[value]; ok {
		return me.declareFnPtr(fn)
	}

	if en, ok := module.enums[value]; ok {
		var gtypes []*datatype
		if me.token.is == "<" {
			gtypes = me.declareGeneric(implementation, en)
			value += genericslist(gtypes)
			if implementation {
				if _, ok := en.module.enums[value]; !ok {
					me.defineEnumImplGeneric(en, value, gtypes)
				}
			}
			if engen, ok := en.module.enums[value]; ok {
				en = engen
			}
		}
		var un *union
		if me.token.is == "." {
			var ok bool
			me.eat(".")
			un, ok = en.types[me.token.value]
			if !ok {
				panic(me.fail() + "Union type \"" + me.token.value + "\" not found for enum \"" + en.name + "\".")
			}
			me.eat("id")
		}
		return newdataenum(me.hmfile, en, un, gtypes)
	}

	if cl, ok := module.classes[value]; ok {
		var gtypes []*datatype
		if me.token.is == "<" {
			gtypes = me.declareGeneric(implementation, cl)
			value += genericslist(gtypes)
			if implementation {
				if _, ok := cl.module.classes[value]; !ok {
					me.defineClassImplGeneric(cl, value, gtypes)
				}
			}
			if clgen, ok := cl.module.classes[value]; ok {
				cl = clgen
			}
		}
		return newdataclass(me.hmfile, cl, gtypes)
	}

	if module != me.hmfile {
		panic(me.fail() + "Unknown declared type \"" + original + "\".")
	}

	// TODO: return newdataunknown(me.hmfile, original, nil)

	return getdatatype(me.hmfile, original)
}

func sizeOfArray(typed string) string {
	i := strings.Index(typed, "]")
	return typed[1:i]
}

func typeOfArrayOrSlice(typed string) (string, string) {
	i := strings.Index(typed, "]")
	size := ""
	if i > 1 {
		size = typed[1:i]
	}
	member := typed[i+1:]
	return size, member
}

func checkIsArrayOrSlice(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '['
}

func checkIsArray(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '[' && typed[1] != ']'
}

func checkIsSlice(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '[' && typed[1] == ']'
}

func checkHasGeneric(typed string) bool {
	return strings.HasSuffix(typed, ">")
}

func checkIsFunction(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '('
}

func (me *parser) assignable(n *node) bool {
	return n.is == "variable" || n.is == "member-variable" || n.is == "array-member"
}

func functionSigType(typed string) ([]string, string) {
	end := strings.Index(typed, ")")

	ret := typed[end+1:]
	ret = strings.TrimSpace(ret)
	if ret == "" {
		ret = "void"
	}

	argd := strings.TrimSpace(typed[1:end])
	args := make([]string, 0)

	if argd != "" {
		for _, a := range strings.Split(argd, ",") {
			args = append(args, strings.TrimSpace(a))
		}
	}

	return args, ret
}
