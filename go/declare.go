package main

import (
	"strings"
)

func (me *parser) declareGeneric(size int) []*datatype {
	me.eat("<")
	order := make([]*datatype, 0)
	for i := 0; i < size; i++ {
		if i != 0 {
			me.eat(",")
		}
		gimpl := me.declareType()
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
			typed := me.declareType()
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
		fn.returns = me.declareType()
	} else {
		fn.returns = newdatavoid()
	}

	return fn.newdatatype()
}

func (me *parser) declareFnPtr(fn *function) *datatype {
	return getdatatype(me.hmfile, fn._name)
}

func (me *parser) declareType() *datatype {

	if me.token.is == "[" {
		me.eat("[")
		if me.token.is == "]" {
			me.eat("]")
			return newdataslice(me.declareType())
		}
		sizeNode := me.calc(0, nil)
		if sizeNode.value == "" || !sizeNode.data().isInt() {
			panic(me.fail() + "array size must be constant integer")
		}
		me.eat("]")
		return newdataarray(sizeNode.value, me.declareType())
	}

	module := me.hmfile
	value := ""

	if me.token.is == "(" {
		return me.declareFn()

	} else if me.token.is == "maybe" {
		me.eat("maybe")
		me.eat("<")
		option := me.declareType()
		me.eat(">")
		return newdatamaybe(option)

	} else if me.token.is == "none" {
		me.eat("none")
		if me.token.is == "<" {
			me.eat("<")
			option := me.declareType()
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
		return newdataany()
	}

	if value == "*" {
		return newdataany()
	}

	if value == TokenString {
		return newdatastring()
	}

	if checkIsPrimitive(value) {
		return newdataprimitive(value)
	}

	if strings.HasPrefix(value, "%") {
		if m, ok := me.hmfile.program.modules[value]; ok && me.token.is == "." {
			me.eat(".")
			module = m
			value = me.token.value
			me.eat("id")
		}
	} else if m, ok := me.hmfile.imports[value]; ok && me.token.is == "." {
		me.eat(".")
		module = m
		value = me.token.value
		me.eat("id")
	}

	if fn, ok := module.functions[value]; ok {
		return me.declareFnPtr(fn)
	}

	if en, ok := module.enums[value]; ok {
		var gtypes []*datatype
		if me.token.is == "<" {
			gtypes = me.declareGeneric(len(en.generics))
			value += genericslist(gtypes)
			if _, ok := en.module.enums[value]; !ok {
				me.defineEnumImplGeneric(en, gtypes)
			}
			if engen, ok := en.module.enums[value]; ok {
				en = engen
			}
		}
		var un *union
		if me.token.is == "." {
			me.eat(".")
			if un = en.getType(me.token.value); un == nil {
				panic(me.fail() + "Union type \"" + me.token.value + "\" not found for enum \"" + en.name + "\".")
			}
			me.eat("id")
		}
		return newdataenum(me.hmfile, en, un, gtypes)
	}

	if cl, ok := module.classes[value]; ok {
		var gtypes []*datatype
		if me.token.is == "<" {
			gtypes = me.declareGeneric(len(cl.generics))
			value += genericslist(gtypes)
			if _, ok := cl.module.classes[value]; !ok {
				me.defineClassImplGeneric(cl, gtypes)
			}
			if clgen, ok := cl.module.classes[value]; ok {
				cl = clgen
			}
		}
		return newdataclass(me.hmfile, cl, gtypes)
	}

	if module != me.hmfile {
		panic(me.fail() + "Unknown declared type \"" + value + "\".")
	}

	return getdatatype(me.hmfile, value)
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
