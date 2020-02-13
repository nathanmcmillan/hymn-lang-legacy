package main

func (me *parser) defaultValue(data *datatype, from string) *node {
	d := nodeInit(data.getRaw())
	d.copyData(data)
	if data.isString() {
		d.value = ""
	} else if data.isChar() {
		d.value = "'\\0'"
	} else if data.isNumber() {
		d.value = "0"
	} else if data.isBoolean() {
		d.value = "false"
	} else if data.isArray() {
		t := nodeInit("array")
		t.copyData(d.data())
		d = t
	} else if data.isSlice() {
		t := nodeInit("slice")
		t.copyData(d.data())
		s := nodeInit(TokenInt)
		s.copyData(getdatatype(me.hmfile, TokenInt))
		s.value = "0"
		t.push(s)
		d = t
	} else if _, ok := data.isClass(); ok {
		t := nodeInit("new")
		t.copyData(d.data())
		me.pushAllDefaultClassParams(t)
		d = t
	} else if data.isSomeOrNone() {
		t := nodeInit("none")
		t.copyData(d.data())
		t.value = "NULL"
		d = t
	} else {
		e := me.fail()
		if from != "" {
			e += "\nFrom: " + from
		}
		e += "\nType: " + d.is + "\nProblem: No default value available."
		panic(e)
	}
	return d
}
