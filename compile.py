functions = dict()
variables = dict()

functions["echo"] = True


def read(tree):
    op = tree["type"]
    if op == "program":
        for s in tree["value"]:
            read(s)
        return None
    if op == "+":
        return read(tree["left"]) + read(tree["right"])
    if op == "-":
        return read(tree["left"]) - read(tree["right"])
    if op == "*":
        return read(tree["left"]) * read(tree["right"])
    if op == "/":
        return read(tree["left"]) / read(tree["right"])
    if op == "number":
        return int(tree["value"])
    if op == "call":
        id = tree["id"]
        val = tree["value"]
        if id == "echo":
            print("echo", read(val))
            return None
    if op == "assign":
        id = tree["id"]
        val = tree["value"]
        variables[id] = read(val)
        return None
    if op == "func":
        id = tree["id"]
        val = tree["value"]
        functions[id] = read(val)
        return None
    if op == "id":
        val = tree["value"]
        if val in functions:
            return functions[val]
        else:
            return variables[val]
    raise AssertionError("compiler unexpected operation \"" + tree + "\"")
