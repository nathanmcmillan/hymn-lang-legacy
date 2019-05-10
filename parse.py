functions = dict()
variables = dict()

functions["echo"] = True


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
    op = token["type"]
    if op == "number" or op == "id":
        eat(parse, op)
        return token
    if op == "(":
        eat(parse, "(")
        tree = calc(parse)
        eat(parse, ")")
        return tree
    raise AssertionError("syntax error in factor @ " + str(parse.pos))


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
    eat(parse, "id")
    tree = dict()
    tree["type"] = "call"
    tree["id"] = token["value"]
    tree["value"] = calc(parse)
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


def defineFunction(parse):
    eat(parse, "func")
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
    tree["type"] = "func"
    tree["args"] = args
    tree["id"] = name
    tree["value"] = calc(parse)
    return tree


def statement(parse):
    token = parse.token
    op = token["type"]
    if op == "id":
        val = token["value"]
        if val in functions:
            tree = call(parse)
        else:
            tree = assign(parse)
    if op == "func":
        tree = defineFunction(parse)
    eat(parse, "line")
    return tree


def program(parse):
    branches = [statement(parse)]
    while parse.token["type"] != "eof":
        branches.append(statement(parse))
    tree = dict()
    tree["type"] = "program"
    tree["value"] = branches
    print(tree)
    return tree


def read(tokens):
    parse = Parser(tokens)
    return program(parse)
