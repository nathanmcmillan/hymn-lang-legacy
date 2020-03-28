package main

import "fmt"

func (me *cfile) compileTry(n *node) *codeblock {

	// # HmEnumUnionResultStringParseError try_0 = example();
	// # if try_0.type == HmResultError {
	// #     return try_0.error;
	// # }
	// # hmstring ex = try_0.ok;

	calc := n.has[0]
	catch := n.has[1]

	cb := &codeblock{}

	temp := "try_" + me.temp()

	d := nodeInit("variable")
	d.idata = newidvariable(me.hmfile, temp)
	d.copyData(calc.data())
	decl := me.compileDeclare(d)

	value := me.eval(calc).code()

	code := decl + " = " + value + me.maybeColon(value) + "\n"
	cn := codeNode(n, code)
	cn.value = temp

	if catch.is == "auto-catch" {
		en, _, _ := catch.data().isEnum()
		lastType := en.types[len(en.types)-1]
		last := enumTypeName(en.baseEnum().cname, lastType.name)
		code := ""
		code += fmc(me.depth) + "if (" + temp + "->type == " + last + ") {\n"
		notTheSame := true
		if notTheSame {
			wrapper := "catch_" + me.temp()
			code += fmc(me.depth+1) + catch.data().typeSig(me) + wrapper + ";\n"
			code += fmc(me.depth+1) + "return " + wrapper + ";\n"
		} else {
			code += fmc(me.depth+1) + "return " + temp + ";\n"
		}
		code += fmc(me.depth) + "}\n"
		cb.prepend(codeBlockOne(n, code))
	}

	cb.prepend(codeNodeUpgrade(cn))

	fmt.Println("[COMPILE CATCH]:", catch.string(me.hmfile, 0))

	en, _, _ := calc.data().isEnum()
	firstType := en.types[0]
	assign := temp + "->" + firstType.name
	cb.current = codeNode(n, assign)

	return cb
}
