functions = dict()
variables = dict()

functions["echo"] = True


def echo(f, space, code):
    s = ""
    for _ in range(space):
        s += "  "
    e = s + code
    print(e)
    f.write(e + "\n")


def write(f, tree, space):
    op = tree["type"]
    if op == "program":
        echo(f, space, "#include <stdio.h>")
        echo(f, space, "")
        echo(f, space, "int main() {")
        for statement in tree["value"]:
            write(f, statement, space + 1)
        echo(f, space + 1, "return 0;")
        echo(f, space, "}")
        return None
    if op == "+":
        return write(f, tree["left"], space) + write(f, tree["right"], space)
    if op == "-":
        return write(f, tree["left"], space) - write(f, tree["right"], space)
    if op == "*":
        return write(f, tree["left"], space) * write(f, tree["right"], space)
    if op == "/":
        return write(f, tree["left"], space) / write(f, tree["right"], space)
    if op == "number":
        return int(tree["value"])
    if op == "call":
        id = tree["id"]
        val = tree["value"]
        if id == "echo":
            echo(f, space, "printf(\"" + str(write(f, val, space)) + "\\n\");")
            return None
    if op == "assign":
        id = tree["id"]
        val = tree["value"]
        variables[id] = write(f, val, space)
        return None
    if op == "func":
        id = tree["id"]
        val = tree["value"]
        functions[id] = write(f, val, space)
        return None
    if op == "id":
        val = tree["value"]
        if val in functions:
            return functions[val]
        else:
            return variables[val]
    raise AssertionError("ccode unexpected operation \"" + tree + "\"")
