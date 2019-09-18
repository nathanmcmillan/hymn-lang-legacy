package main

import "fmt"

func (me *cfile) walrusIf(n *node) string {
	ifnode := n.has[0]
	has := len(ifnode.has)
	code := ""
	if ifnode.is == ":=" {
		delete(ifnode.attributes, "parenthesis")
		code += me.eval(ifnode).code + ";\n" + fmc(me.depth)
		n.has[0] = ifnode.has[0]
	} else if has > 0 {
		ifleft := ifnode.has[0]
		if ifleft.is == ":=" {
			delete(ifleft.attributes, "parenthesis")
			code += me.eval(ifleft).code + ";\n" + fmc(me.depth)
			ifnode.has[0] = ifleft.has[0]
		} else if has > 1 {
			ifright := ifnode.has[1]
			if ifright.is == ":=" {
				delete(ifright.attributes, "parenthesis")
				code += me.eval(ifright).code + ";\n" + fmc(me.depth)
				ifnode.has[1] = ifright.has[0]
			}
		}
	}
	return code
}

func (me *cfile) walrusLoop(n *node) string {
	ifnode := n.has[0]
	has := len(ifnode.has)
	code := ""
	if ifnode.is == ":=" {
		ifnode.has[0].attributes["mutable"] = "true"
		code += me.declare(ifnode.has[0]) + ";\n" + fmc(me.depth)
	}
	if has > 0 {
		ifleft := ifnode.has[0]
		if ifleft.is == ":=" {
			ifleft.has[0].attributes["mutable"] = "true"
			code += me.declare(ifleft.has[0]) + ";\n" + fmc(me.depth)
		}
		if has > 1 {
			ifright := ifnode.has[1]
			if ifright.is == ":=" {
				ifright.has[0].attributes["mutable"] = "true"
				code += me.declare(ifright.has[0]) + ";\n" + fmc(me.depth)
			}
		}
	}
	return code
}

func (me *cfile) walrusMatch(n *node) string {
	ifnode := n.has[0]
	code := ""
	fmt.Println("WALRUS MATCH ::", n.string(0))
	if ifnode.is == ":=" {
		delete(ifnode.attributes, "parenthesis")
		code += me.eval(ifnode).code + ";\n" + fmc(me.depth)
		n.has[0] = ifnode.has[0]
	}
	return code
}
