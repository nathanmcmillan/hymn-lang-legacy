class Parser {
    static read(str) {
        let data = {}
        let stack = [data]
        let key = ""
        let value = ""
        let state = "key"
        for (let i = 0; i < str.length; i++) {
            let c = str[i]
            if (c === ":") {
                state = "value"
            } else if (c === ",") {
                let pc = str[i - 1]
                if (pc !== "}" && pc !== "]") {
                    if (stack[0].constructor === Array) {
                        stack[0].push(value)
                    } else {
                        stack[0][key] = value
                        key = ""
                        state = "key"
                    }
                    value = ""
                }
            } else if (c === "{") {
                let map = {}
                if (stack[0].constructor === Array) {
                    stack[0].push(map)
                    state = "key"
                } else {
                    stack[0][key] = map
                    key = ""
                }
                stack.unshift(map)
            } else if (c === "[") {
                let array = []
                if (stack[0].constructor === Array) {
                    stack[0].push(array)
                } else {
                    stack[0][key] = array
                    key = ""
                }
                stack.unshift(array)
                state = "value"
            } else if (c === "}") {
                let pc = str[i - 1]
                if (pc !== "," && pc !== "{" && pc !== "]" && pc !== "}") {
                    stack[0][key] = value
                    key = ""
                    value = ""
                }
                stack.shift()
                if (stack[0].constructor === Array) state = "value"
                else state = "key"
            } else if (c === "]") {
                let pc = str[i - 1]
                if (pc !== "," && pc !== "[" && pc !== "]" && pc !== "}") {
                    stack[0].push(value)
                    value = ""
                }
                stack.shift()
                if (stack[0].constructor === Array) state = "value"
                else state = "key"
            } else if (state === "key") {
                key += c
            } else
                value += c
        }
        let pc = str[str.length - 1]
        if (pc !== "," && pc !== "]" && pc !== "}")
            stack[0][key] = value
        return data
    }
}
