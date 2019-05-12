functions = dict()
variables = dict()

functions["echo"] = {
    "args": ["s"]
}


class Parser():
    def __init__(self, tokens):
        self.tokens = tokens
        self.token = tokens[0]
        self.pos = 0

    def next(self):
        self.pos += 1
        self.token = self.tokens[self.pos]

    def peek(self):
        return self.tokens[self.pos]


def eat(parse, expecting):
    token = parse.token
    op = token["type"]
    if op != expecting:
        raise AssertionError("token was \"" + op + "\" instead of \"" + expecting + "\" @ " + str(parse.pos))
    parse.next()


def binary(left, op, right):
    tree = dict()
    tree["left"] = left
    tree["type"] = op
    tree["right"] = right
    return tree


def factor(parse):
    token = parse.token
    print("(factor op)", token)
    op = token["type"]
    if op == "number":
        eat(parse, op)
        return token
    if op == "id":
        name = token["value"]
        if name in functions:
            return call(parse)
        else:
            eat(parse, op)
            return token
    if op == "(":
        eat(parse, "(")
        tree = calc(parse)
        eat(parse, ")")
        return tree
    raise AssertionError("factor error in factor @ " + str(parse.pos))


def term(parse):
    tree = factor(parse)
    while True:
        token = parse.token
        op = token["type"]
        if op == "*" or op == "/":
            eat(parse, op)
            tree = binary(tree, op, factor(parse))
            continue
        break
    return tree


def calc(parse):
    tree = term(parse)
    while True:
        token = parse.token
        op = token["type"]
        if op == "+" or op == "-":
            eat(parse, op)
            tree = binary(tree, op, term(parse))
            continue
        break
    return tree


def call(parse):
    token = parse.token
    print("(call op)")
    name = token["value"]
    args = functions[name]["args"]
    eat(parse, "id")
    tree = dict()
    tree["type"] = "call"
    tree["func"] = name
    parameters = []
    for _ in args:
        parameters.append(calc(parse))
    tree["parameters"] = parameters
    return tree


def assign(parse):
    token = parse.token
    eat(parse, "id")
    eat(parse, "=")
    tree = dict()
    tree["type"] = "assign"
    tree["id"] = token["value"]
    tree["value"] = calc(parse)
    return tree


def functionOp(parse):
    eat(parse, "function")
    token = parse.token
    name = token["value"]
    if name in functions:
        raise AssertionError("function \"" + name + "\" already defined @ " + str(parse.pos))
    eat(parse, "id")
    args = []
    while parse.token["type"] != "line":
        args.append(parse.token["value"])
        eat(parse, "id")
    eat(parse, "line")
    tree = dict()
    tree["args"] = args
    branches = []
    while True:
        token = parse.token
        if token["type"] == "line":
            eat(parse, "line")
            break
        if token["type"] == "eof":
            break
        leaf = statement(parse)
        branches.append(leaf)
        if leaf["type"] == "return":
            break
    tree["value"] = branches
    functions[name] = tree
    return tree


def returnOp(parse):
    eat(parse, "return")
    tree = dict()
    tree["type"] = "return"
    print("(return op)")
    tree["value"] = calc(parse)
    return tree


def statement(parse):
    token = parse.token
    op = token["type"]
    if op == "id":
        name = token["value"]
        if name in functions:
            tree = call(parse)
        else:
            tree = assign(parse)
        eat(parse, "line")
    elif op == "function":
        functionOp(parse)
        tree = None
    elif op == "return":
        tree = returnOp(parse)
        eat(parse, "line")
    elif op == "line":
        tree = None
        eat(parse, "line")
    elif op == "eof":
        return None
    else:
        raise AssertionError("unknown statement op @ " + str(parse.pos))
    return tree


def program(parse):
    while parse.token["type"] != "eof":
        statement(parse)

    prog = dict()
    prog["imports"] = dict()
    prog["globals"] = dict()
    prog["locals"] = dict()
    prog["functions"] = functions

    print(prog)
    return prog


def read(tokens):
    parse = Parser(tokens)
    return program(parse)
