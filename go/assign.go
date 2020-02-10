package main

import (
	"fmt"
	"strings"
)

func (me *parser) assign(left *node, malloc, mutable bool) *node {
	op := me.token.is
	mustBeInt := false
	mustBeNumber := false
	if op == "%=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		mustBeInt = true
	} else if op == "-=" || op == "*=" || op == "/=" {
		mustBeNumber = true
	} else if op != "+=" && op != "=" && op != ":=" {
		if op == "+" {
			panic(me.fail() + "Use \"+= 1\" rather than \"++\".")
		}
		panic(me.fail() + "Unknown assign operation \"" + op + "\".")
	}
	me.eat(op)
	right := me.calc(0, left.data())
	if mustBeInt {
		if !right.data().isInt() {
			panic(me.fail() + "assign operation \"" + op + "\" requires int type")
		}
	} else if mustBeNumber {
		if !right.data().isNumber() {
			panic(me.fail() + "assign operation \"" + op + "\" requires number type")
		}
	}
	if left.is == "variable" {
		sv := me.hmfile.getvar(left.idata.name)
		if sv != nil {
			if !sv.mutable {
				panic(me.fail() + "Variable: " + sv.name + " is not mutable.")
			}
			if !right.data().isQuestion() && left.data().notEquals(right.data()) {
				enleft, _, ok1 := left.data().isEnum()
				enright, _, ok2 := right.data().isEnum()
				if ok1 && ok2 && enleft == enright {
					left.copyDataOfNode(right)
				} else if strings.HasPrefix(left.data().getRaw(), right.data().getRaw()) && strings.Index(left.data().getRaw(), "<") != -1 {
					right.copyDataOfNode(left)
				} else {
					panic(me.fail() + "Variable: " + sv.name + " of type: " + left.data().print() + " does not match expression: " + right.data().print())
				}
			}
		} else if mustBeInt || mustBeNumber || op == "+=" {
			panic(me.fail() + "cannot operate \"" + op + "\" because variable \"" + left.idata.name + "\" does not exist.")
		} else {
			if mutable {
				left.attributes["mutable"] = "true"
			}
			if !malloc {
				right.data().setIsPointer(false)
			}
			left.copyDataOfNode(right)
			varini := right.data().getnamedvariable(left.idata.name, mutable)
			me.hmfile.scope.variables[left.idata.name] = varini

			if mutable && left.idata.name == "e" {
				fmt.Println("DEBUG ::", left.data().error())
			}

		}
	} else if left.is == "member-variable" || left.is == "array-member" {
		if !right.data().isQuestion() && left.data().notEquals(right.data()) {
			if strings.HasPrefix(left.data().getRaw(), right.data().getRaw()) && strings.Index(left.data().getRaw(), "<") != -1 {
				right.copyDataOfNode(left)
			} else {
				panic(me.fail() + "Member variable: " + left.data().error() + " does not match expression: " + right.data().error())
			}
		}
	} else {
		panic(me.fail() + "bad assignment \"" + left.is + "\"")
	}
	if left.idata != nil && left.is == "variable" {
		right.attributes["assign"] = left.idata.name
	}
	if _, useStack := right.attributes["stack"]; useStack {
		left.attributes["stack"] = "true"
	}
	n := nodeInit(op)
	if op == ":=" {
		n.copyDataOfNode(right)
	}
	n.push(left)
	n.push(right)
	return n
}

func (me *parser) forceassign(v *node, malloc, mutable bool) *node {
	if !me.assignable(v) {
		panic(me.fail() + "Expected variable for assignment but was \"" + v.data().print() + "\".")
	}
	return me.assign(v, malloc, mutable)
}
