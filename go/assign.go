package main

import (
	"strings"
)

func (me *parser) assign(left *node, malloc, mutable bool) (*node, *parseError) {
	op := me.token.is
	mustBeInt := false
	mustBeNumber := false
	if op == "%=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		mustBeInt = true
	} else if op == "-=" || op == "*=" || op == "/=" {
		mustBeNumber = true
	} else if op != "+=" && op != "=" && op != ":=" {
		if op == "+" {
			return nil, err(me, ECodeDoublePlus, "Use \"+= 1\" rather than \"++\".")
		}
		return nil, err(me, ECodeUnknownAssignOperation, "Unknown assign operation \""+op+"\".")
	}
	if er := me.eat(op); er != nil {
		return nil, er
	}
	right, er := me.calc(0, left.data())
	if er != nil {
		return nil, er
	}
	if mustBeInt {
		if !right.data().isInt() {
			return nil, err(me, ECodeAssignOperationRequiresInteger, "assign operation \""+op+"\" requires int type")
		}
	} else if mustBeNumber {
		if !right.data().isNumber() {
			return nil, err(me, ECodeAssignOperationRequiresNumber, "assign operation \""+op+"\" requires number type")
		}
	}
	if left.is == "variable" {
		sv := me.hmfile.getvar(left.idata.name)
		if sv != nil {
			if !sv.mutable {
				return nil, err(me, ECodeVariableNotMutable, "Variable: "+sv.name+" is not mutable.")
			}
			if !right.data().isAnyType() && left.data().notEquals(right.data()) {
				enleft, _, ok1 := left.data().isEnum()
				enright, _, ok2 := right.data().isEnum()
				if ok1 && ok2 && enleft == enright {
					left.copyDataOfNode(right)
				} else if strings.HasPrefix(left.data().getRaw(), right.data().getRaw()) && strings.Index(left.data().getRaw(), "<") != -1 {
					right.copyDataOfNode(left)
				} else {
					return nil, err(me, ECodeVariableTypeMismatch, "Variable `"+sv.name+"` of type `"+left.data().print()+"` does not match expression `"+right.data().print()+"`")
				}
			}
		} else if mustBeInt || mustBeNumber || op == "+=" {
			return nil, err(me, ECodeVariableDoesNotExist, "cannot operate \""+op+"\" because variable \""+left.idata.name+"\" does not exist.")
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
		}
	} else if left.is == "member-variable" || left.is == "array-member" {
		if !right.data().isAnyType() && left.data().notEquals(right.data()) {
			if strings.HasPrefix(left.data().getRaw(), right.data().getRaw()) && strings.Index(left.data().getRaw(), "<") != -1 {
				right.copyDataOfNode(left)
			} else {
				return nil, err(me, ECodeMemberVariableTypeMismatch, "Member variable: "+left.data().error()+" does not match expression: "+right.data().error())
			}
		}
	} else {
		return nil, err(me, ECodeBadAssignment, "bad assignment \""+left.is+"\"")
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
	return n, nil
}

func (me *parser) forceassign(v *node, malloc, mutable bool) (*node, *parseError) {
	if !me.assignable(v) {
		return nil, err(me, ECodeExpectedToAssignVariable, "Expected variable for assignment but was \""+v.data().print()+"\".")
	}
	return me.assign(v, malloc, mutable)
}
