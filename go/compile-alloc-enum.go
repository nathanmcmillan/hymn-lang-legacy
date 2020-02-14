package main

func (me *cfile) compileAllocEnum(n *node) *codeblock {

	data := n.data()

	en, un, _ := data.isEnum()
	enumType := un.name
	if en.simple {
		enumBase := en.cname
		globalName := enumTypeName(enumBase, enumType)
		return codeBlockOne(n, globalName)
	}

	baseEnumName := en.baseEnum().cname
	unionName := en.ucname

	_, useStack := n.attributes["stack"]
	assignName, local := n.attributes["assign"]
	var assign *variable
	if local {
		assign = me.getvar(assignName)
	}

	cb := &codeblock{}

	var memberRef string
	if useStack {
		memberRef = "."
	} else {
		memberRef = "->"
	}

	if local {
		code := ""
		if !useStack {
			code += "malloc(sizeof(" + unionName + "))"
		}
		code += ";\n" + fmc(me.depth) + assign.cname + memberRef + "type = " + enumTypeName(baseEnumName, un.name)
		xvar := un.types.size() > 1
		for i, typeKey := range un.types.order {
			v := un.types.table[typeKey]
			p := n.has[i]
			if !v.isPointer() {
				p.attributes["stack"] = "true"
			}
			cassign := ";\n" + fmc(me.depth) + assign.cname + memberRef + un.name
			if xvar {
				cassign += "." + typeKey
			}
			cassign += " = "
			if p.is == "new" || (p.is == "enum" && p.data().union != nil) {
				temp := me.temp()
				p.attributes["assign"] = temp
				d := nodeInit("variable")
				d.idata = newidvariable(me.hmfile, temp)
				d.copyDataOfNode(p)
				decl := me.compileDeclare(d)
				value := me.eval(p).code()
				code2 := ";\n" + fmc(me.depth) + decl
				if v.isPointer() {
					code2 += " = "
				}
				code2 += value
				cn := codeNode(p, code2)
				cn.value = temp
				code += cn.code
				code += cassign + cn.value
			} else {
				code += cassign + me.eval(p).code()
			}
		}
		cb.current = codeNode(n, code)

	} else {
		temp := me.temp()
		cb.current = codeNode(n, temp)
		n.attributes["assign"] = temp
		d := nodeInit("variable")
		d.idata = newidvariable(me.hmfile, temp)
		d.copyDataOfNode(n)
		decl := me.compileDeclare(d)
		value := me.eval(n).code()
		code := decl + " = " + value + me.maybeColon(value) + "\n"
		cn := codeNode(n, code)
		cn.value = temp
		cb.prepend(codeNodeUpgrade(cn))
	}

	return cb
}
