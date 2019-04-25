const fs = require("fs")

sigil(process.argv[2])

function sigil(file) {
    console.log("file: " + file)
    fs.readFile(file, "utf8", (err, data) => {
        if (err) {
            throw err
        }
        compile(data)
    })
}

const TokenSet = "set"
const TokenConst = "constant"
const TokenFn = "function"
const TokenEnd = "end"
const TokenScopeStart = "("
const TokenScopeEnd = ")"
const TokenWord = "word"
const TokenUnknown = "?"
const TokenNewLine = "line"

const tokenMap = new Map()
tokenMap.set("set", TokenSet)
tokenMap.set("const", TokenConst)
tokenMap.set("def", TokenFn)
tokenMap.set("--", TokenEnd)
tokenMap.set("(", TokenScopeStart)
tokenMap.set(")", TokenScopeEnd)
tokenMap.set("\n", TokenNewLine)

const SyntaxLeftScope = "("
const SyntaxRightScope = ")"

class Scope {
    constructor() {
        this.mutables = new Map()
        this.immutables = new Map()
        this.defs = new Map()
    }
}

const global = new Scope()


function patternWord(word) {
    return /^[A-z0-9]+$/.test(word)
}

class Token {
    constructor(type, value) {
        this.type = type
        this.value = value
    }
    static typeOf(word) {
        let type = tokenMap.get(word)
        if (type) {
            return type
        }
        if (patternWord(word)) {
            return TokenWord
        }
        return TokenUnknown
    }
    static find(value) {
        let type = Token.typeOf(value)
        return new Token(type, value)
    }
}

function compile(data) {
    let tokens = new Tokenize().process(data)
    console.log("===")
    tokens.forEach((v) => {
        console.log(v)
    })
    console.log("===")
}

class Tokenize {
    constructor() {
        this.tokens = new Array()
        this.word = ""
        this.define = "?"
    }
    stop() {
        if (this.word == "") {
            return
        }
        this.tokens.push(Token.find(this.word))
        this.word = ""
    }
    process(data) {
        for (let i = 0; i < data.length; i++) {
            let c = data[i]
            if (c == "\n") {
                this.stop()
                this.tokens.push(new Token(TokenNewLine, c))
            } else if (c == " " || c == "\t") {
                this.stop()
            } else if (patternWord(c)) {
                if (this.define != "word") {
                    this.stop()
                }
                this.word += c
                this.define = "word"
            } else {
                if (this.define != "?") {
                    this.stop()
                }
                this.word += c
                this.define = "?"
            }
        }
        return this.tokens
    }
}
