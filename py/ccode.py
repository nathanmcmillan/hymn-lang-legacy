class CodeRef():
    def __init__(self):
        self.code = ""

    def append(self, code):
        self.code += code


class CCode():
    def __init__(self, functions, function_order):
        self.functions = functions
        self.function_order = function_order
        self.funcs = dict()

    def build(self):
        head = ""
        head += "#include <stdio.h>\n"
        head += "\n"

        code = ""
        code += head

        for func in self.function_order:
            if func in self.funcs:
                code += self.funcs[func].code

        return code


def fmt(space, code):
    s = ""
    for _ in range(space):
        s += "  "
    return s + code


def callOp(name, parameters, variables):
    if name == "echo":
        param = emit(None, parameters[0], 0, variables)
        param_type = param["type"]
        param_value = param["value"]
        if param_type == "string":
            s = "printf(\"%s\\n\", \"" + param_value + "\");"
        elif param_type == "number":
            s = "printf(\"%d\\n\", " + param_value + ");"
        else:
            raise Exception("bad argument to echo")
    else:
        s = name + "("
        first = True
        for p in parameters:
            if first:
                first = False
            else:
                s += ", "
            s += emit(None, p, 0, variables)
        s += ")"
    return s


def objectOp(name, fields):
    code = []
    code.append("struct " + name + " {")
    for f in fields:
        code.append("int " + f + ";")
    code.append("};")
    return code


def objectInitOp(objectType, name):
    print("new object")
    s = format("struct {} *{} = (struct {} *)malloc(sizeof(struct {}));", objectType, name, objectType, objectType)
    return s


def deleteOp(name):
    print("delete")
    s = "free(" + name + ");"
    return s


def funcOp(ccode, name, tree, space):
    print("func", name, tree)
    args = tree["args"]
    code = CodeRef()
    variables = dict()
    ccode.funcs[name] = code
    d = "int " + name + "("
    first = True
    for argname in args:
        if first:
            first = False
        else:
            d += ", "
        d += "int " + argname
        variables[argname] = dict()
    d += ") {"
    code.append(fmt(space, d))
    for statement in tree["value"]:
        code.append(fmt(space + 1, emit(main, statement, space, variables)))
    code.append(fmt(space, "}"))
    code.append("")


def programOp(ccode, tree, space):
    print("program")
    code = CodeRef()
    variables = dict()
    ccode.funcs["main"] = code
    code.append("int main()\n{\n")
    for statement in tree["value"]:
        emit(code, statement, space, variables)
        # code.append(fmt(space + 1, ))
    code.append(fmt(space + 1, "return 0;\n"))
    code.append("}\n")


def emit(code, tree, space, variables):
    print("tree", tree)
    op = tree["type"]
    if op == "+":
        emit(code, tree["left"], space, variables)
        code.append(" + ")
        emit(code, tree["right"], space, variables)
        return
    if op == "-":
        return emit(code, tree["left"], space, variables)["value"] + " - " + emit(code, tree["right"], space, variables)["value"]
    if op == "*":
        return emit(code, tree["left"], space, variables)["value"] + " * " + emit(code, tree["right"], space, variables)["value"]
    if op == "/":
        return emit(code, tree["left"], space, variables)["value"] + " / " + emit(code, tree["right"], space, variables)["value"]
    if op in ("number", "string"):
        print("primitive")
        return tree
    if op == "call":
        print("call")
        src = callOp(tree["func"], tree["parameters"], variables) + "\n"
        code.append(fmt(space + 1, src))
        return
    if op == "assign":
        print("assign")
        name = tree["id"]
        val = tree["value"]
        s = name + " = " + emit(code, val, space, variables) + ";"
        if name not in variables:
            s = "int " + s
            variables[name] = dict()
        return s
    if op == "return":
        print("return")
        val = tree["value"]
        s = "return " + emit(code, val, space, variables) + ";"
        return s
    if op == "id":
        print("id")
        return tree["value"]
    raise AssertionError("ccode unexpected operation", tree)


def main(cfile, prog, space):
    functions = prog["functions"]
    function_order = prog["function_order"]
    main = functions["main"]
    del functions["main"]
    del functions["echo"]
    ccode = CCode(functions, function_order)
    for func in functions:
        funcOp(ccode, func, functions[func], space)
    programOp(ccode, main, space)
    code = ccode.build()
    print("=== main.c ===")
    print(code)
    with open(cfile, "w") as f:
        f.write(code)
