class CCode():
    def __init__(self, functions):
        self.functions = functions
        self.funcs = dict()

    def build(self):
        head = []
        head.append("#include <stdio.h>")
        head.append("")

        code = []

        for line in head:
            code.append(line)

        for func in self.funcs:
            for line in self.funcs[func]:
                code.append(line)

        return code


def fmt(space, code):
    s = ""
    for _ in range(space):
        s += "  "
    return s + code


def callOp(name, parameters):
    if name == "echo":
        params = read(None, parameters[0], 0)
        s = "printf(\"%d\\n\", " + params + ");"
    else:
        s = name + "("
        first = True
        for p in parameters:
            if first:
                first = False
            else:
                s += ", "
            s += read(None, p, 0)
        s += ")"
    return s


def funcOp(ccode, name, tree, space):
    print("(compile read) func", name, tree)
    args = tree["args"]
    code = []
    ccode.funcs[name] = code
    d = "int " + name + "("
    first = True
    for argname in args:
        if first:
            first = False
        else:
            d += ", "
        d += "int " + argname
    d += ") {"
    code.append(fmt(space, d))
    for statement in tree["value"]:
        code.append(fmt(space + 1, read(main, statement, space)))
    code.append(fmt(space, "}"))
    code.append("")


def programOp(ccode, tree, space):
    # TODO scope
    print("(compile read) program")
    code = []
    ccode.funcs["main"] = code
    code.append(fmt(space, "int main() {"))
    for statement in tree["value"]:
        code.append(fmt(space + 1, read(code, statement, space)))
    code.append(fmt(space + 1, "return 0;"))
    code.append(fmt(space, "}"))


def read(code, tree, space):
    print("(compile read) tree", tree)
    op = tree["type"]
    if op == "+":
        return read(code, tree["left"], space) + " + " + read(code, tree["right"], space)
    if op == "-":
        return read(code, tree["left"], space) + " - " + read(code, tree["right"], space)
    if op == "*":
        return read(code, tree["left"], space) + " * " + read(code, tree["right"], space)
    if op == "/":
        return read(code, tree["left"], space) + " / " + read(code, tree["right"], space)
    if op == "number":
        print("(compile read) number")
        return tree["value"]
    if op == "call":
        print("(compile read) call")
        return callOp(tree["func"], tree["parameters"])
    if op == "assign":
        print("(compile read) assign")
        name = tree["id"]
        val = tree["value"]
        s = "int " + name + " = " + read(code, val, space) + ";"
        return s
    if op == "return":
        print("(compile read) return")
        val = tree["value"]
        s = "return " + read(code, val, space) + ";"
        return s
    if op == "id":
        print("(compile read) id")
        return tree["value"]
    raise AssertionError("ccode unexpected operation", tree)


def main(cfile, prog, space):
    functions = prog["functions"]
    main = functions["main"]
    del functions["main"]
    del functions["echo"]
    ccode = CCode(functions)
    for func in functions:
        funcOp(ccode, func, functions[func], space)
    programOp(ccode, main, space)
    code = ccode.build()
    for line in code:
        print(line)
    with open(cfile, "w") as f:
        for line in code:
            f.write(line + "\n")
