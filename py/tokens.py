def whileSpaces(pos, source):
    size = len(source)
    while pos < size:
        c = source[pos]
        if c == " " or c == "\t":
            pos += 1
            continue
        return pos


def whileNumber(pos, source):
    number = ""
    size = len(source)
    while pos < size:
        c = source[pos]
        if c in "01234567890":
            number += c
            pos += 1
            continue
        break
    return pos, number


def whileWord(pos, source):
    word = ""
    size = len(source)
    while pos < size:
        c = source[pos]
        if c in "abcdefghijklmnopqrstuvwxyz":
            word += c
            pos += 1
            continue
        break
    return pos, word


def whileString(pos, source):
    value = ""
    size = len(source)
    while pos < size:
        c = source[pos]
        pos += 1
        if c != "\"":
            value += c
            continue
        break
    return pos, value


def simpleToken(of):
    token = dict()
    token["type"] = of
    return token


def valueToken(of, value):
    token = dict()
    token["type"] = of
    token["value"] = value
    return token


def peek(pos, source):
    size = len(source)
    if pos < size:
        return source[pos]
    return "eof"


def read(source):
    tokens = []
    pos = 0
    size = len(source)
    while pos < size:
        pos = whileSpaces(pos, source)
        (pos, number) = whileNumber(pos, source)
        if number != "":
            tokens.append(valueToken("number", number))
            continue
        (pos, word) = whileWord(pos, source)
        if word != "":
            if word in ("function", "return", "object", "new", "delete"):
                tokens.append(simpleToken(word))
            else:
                tokens.append(valueToken("id", word))
            continue
        c = source[pos]
        if c == "\"":
            pos += 1
            (pos, value) = whileString(pos, source)
            tokens.append(valueToken("string", value))
            continue
        if c in "+-*/()=.":
            tokens.append(simpleToken(c))
            pos += 1
            continue
        if c == "\n":
            tokens.append(simpleToken("line"))
            pos += 1
            continue
        raise AssertionError("unknown token \"" + c + "\"")
    tokens.append(simpleToken("eof"))
    return tokens
