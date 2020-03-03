package main

import (
	"strings"
)

func (me *parser) declareGeneric(size int) ([]*datatype, *parseError) {
	me.eat("<")
	order := make([]*datatype, 0)
	for i := 0; i < size; i++ {
		if i != 0 {
			me.eat(",")
		}
		gimpl, er := me.declareType()
		if er != nil {
			return nil, er
		}
		order = append(order, gimpl)
	}
	me.eat(">")
	return order, nil
}

func (me *parser) declareFn() (*datatype, *parseError) {
	me.eat("(")
	fn := fnSigInit(me.hmfile)
	if me.token.is != ")" {
		for {
			typed, er := me.declareType()
			if er != nil {
				return nil, er
			}
			fn.args = append(fn.args, fnArgInit(typed.getvariable()))
			if me.token.is == ")" {
				break
			} else if me.token.is == "," {
				me.eat(",")
				continue
			}
			return nil, err(me, ECodeUnexpectedToken, "unexpected token in function pointer")
		}
	}
	me.eat(")")
	if me.token.is != "line" && me.token.is != "," {
		var er *parseError
		fn.returns, er = me.declareType()
		if er != nil {
			return nil, er
		}
	} else {
		fn.returns = newdatavoid()
	}

	return fn.newdatatype()
}

func (me *parser) declareFnPtr(fn *function) (*datatype, *parseError) {
	return getdatatype(me.hmfile, fn._name)
}

func (me *parser) declareType() (*datatype, *parseError) {

	if me.token.is == "[" {
		me.eat("[")
		if me.token.is == "]" {
			me.eat("]")
			decl, er := me.declareType()
			if er != nil {
				return nil, er
			}
			return newdataslice(decl), nil
		}
		sizeNode, er := me.calc(0, nil)
		if er != nil {
			return nil, er
		}
		if sizeNode.value == "" || !sizeNode.data().isInt() {
			return nil, err(me, ECodeArraySizeRequiresInteger, "array size must be constant integer")
		}
		me.eat("]")
		decl, er := me.declareType()
		if er != nil {
			return nil, er
		}
		return newdataarray(sizeNode.value, decl), nil
	}

	module := me.hmfile
	value := ""

	if me.token.is == "(" {
		return me.declareFn()

	} else if me.token.is == "maybe" {
		me.eat("maybe")
		me.eat("<")
		option, er := me.declareType()
		if er != nil {
			return nil, er
		}
		me.eat(">")
		if !option.isPointer() {
			return nil, err(me, ECodeMaybeTypeRequiresPointer, "Maybe type requires a pointer. Found: "+option.print())
		}
		return newdatamaybe(option), nil

	} else if me.token.is == "none" {
		me.eat("none")
		if me.token.is == "<" {
			me.eat("<")
			option, er := me.declareType()
			if er != nil {
				return nil, er
			}
			me.eat(">")
			if !option.isPointer() {
				return nil, err(me, ECodeNoneTypeRequiresPointer, "None type requires a pointer. Found: "+option.print())
			}
			return newdatamaybe(option), nil
		}
		return newdatanone(), nil
	} else {
		value += me.token.value
		me.wordOrPrimitive()
	}

	if value == "void" {
		return newdatavoid(), nil
	}

	if value == "?" {
		return newdataany(), nil
	}

	if value == "*" {
		return newdataany(), nil
	}

	if value == TokenString {
		return newdatastring(), nil
	}

	if checkIsPrimitive(value) {
		return newdataprimitive(value), nil
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
			var er *parseError
			gtypes, er = me.declareGeneric(len(en.generics))
			if er != nil {
				return nil, er
			}
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
				return nil, err(me, ECodeEnumDoesNotHaveType, "Union type \""+me.token.value+"\" not found for enum \""+en.name+"\".")
			}
			me.eat("id")
		}
		return newdataenum(me.hmfile, en, un, gtypes), nil
	}

	if cl, ok := module.classes[value]; ok {
		var gtypes []*datatype
		if me.token.is == "<" {
			var er *parseError
			gtypes, er = me.declareGeneric(len(cl.generics))
			if er != nil {
				return nil, er
			}
			value += genericslist(gtypes)
			if _, ok := cl.module.classes[value]; !ok {
				me.defineClassImplGeneric(cl, gtypes)
			}
			if clgen, ok := cl.module.classes[value]; ok {
				cl = clgen
			}
		}
		return newdataclass(me.hmfile, cl, gtypes), nil
	}

	if module != me.hmfile {
		return nil, err(me, ECodeUnknownType, "Unknown declared type \""+value+"\".")
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
