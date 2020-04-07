package main

type prefixRule struct {
	precedence int
	name       string
	fn         func(*parser, string) (*node, *parseError)
}

type infixRule struct {
	precedence int
	name       string
	fn         func(*parser, *node, string) (*node, *parseError)
}

var (
	prefixes map[string]prefixRule
	infixes  map[string]infixRule

	operators = map[string]bool{
		"=":   true,
		":=":  true,
		"+=":  true,
		"-=":  true,
		"*=":  true,
		"/=":  true,
		"%=":  true,
		"&=":  true,
		"|=":  true,
		"^=":  true,
		"<<=": true,
		">>=": true,
	}
)

func init() {
	prefixes = map[string]prefixRule{
		TokenIntLiteral:     prefixRule{6, "", prefixPrimitive},
		TokenFloatLiteral:   prefixRule{6, "", prefixPrimitive},
		TokenBooleanLiteral: prefixRule{6, "", prefixPrimitive},
		TokenStringLiteral:  prefixRule{6, "", prefixString},
		TokenCharLiteral:    prefixRule{6, "", prefixChar},
		"none":              prefixRule{6, "", prefixNone},
		"maybe":             prefixRule{6, "", prefixMaybe},
		"id":                prefixRule{6, "", prefixIdent},
		"$":                 prefixRule{8, "", prefixIdent},
		"+":                 prefixRule{8, "+sign", prefixSign},
		"-":                 prefixRule{8, "-sign", prefixSign},
		"!":                 prefixRule{8, "not", prefixNot},
		TokenSizeT:          prefixRule{8, "cast", prefixCast},
		TokenInt:            prefixRule{8, "cast", prefixCast},
		TokenInt8:           prefixRule{8, "cast", prefixCast},
		TokenInt16:          prefixRule{8, "cast", prefixCast},
		TokenInt32:          prefixRule{8, "cast", prefixCast},
		TokenInt64:          prefixRule{8, "cast", prefixCast},
		TokenUInt:           prefixRule{8, "cast", prefixCast},
		TokenUInt8:          prefixRule{8, "cast", prefixCast},
		TokenUInt16:         prefixRule{8, "cast", prefixCast},
		TokenUInt32:         prefixRule{8, "cast", prefixCast},
		TokenUInt64:         prefixRule{8, "cast", prefixCast},
		TokenFloat:          prefixRule{8, "cast", prefixCast},
		TokenFloat32:        prefixRule{8, "cast", prefixCast},
		TokenFloat64:        prefixRule{8, "cast", prefixCast},
		"not":               prefixRule{8, "", prefixNot},
		"[":                 prefixRule{9, "", prefixArray},
		"(":                 prefixRule{10, "", prefixGroup},
	}

	infixes = map[string]infixRule{
		":=":  infixRule{1, "", infixWalrus},
		"?":   infixRule{1, "", infixTernary},
		"and": infixRule{1, "", infixCompare},
		"or":  infixRule{1, "", infixCompare},
		">":   infixRule{2, "", infixCompare},
		">=":  infixRule{2, "", infixCompare},
		"<":   infixRule{2, "", infixCompare},
		"<=":  infixRule{2, "", infixCompare},
		"==":  infixRule{2, "equal", infixCompare},
		"!=":  infixRule{2, "not-equal", infixCompare},
		"is":  infixRule{2, "", infixCompareEnumIs},
		">>":  infixRule{2, "", infixBinaryInt},
		"<<":  infixRule{2, "", infixBinaryInt},
		"&":   infixRule{2, "", infixBinaryInt},
		"|":   infixRule{2, "", infixBinaryInt},
		"^":   infixRule{2, "", infixBinaryInt},
		"%":   infixRule{4, "", infixBinaryInt},
		"+":   infixRule{3, "", infixBinary},
		"-":   infixRule{3, "", infixBinary},
		"*":   infixRule{4, "", infixBinary},
		"/":   infixRule{4, "", infixBinary},
	}
}

func getPrefixPrecedence(op string) int {
	if pre, ok := prefixes[op]; ok {
		return pre.precedence
	}
	return 0
}

func getInfixPrecedence(op string) int {
	if inf, ok := infixes[op]; ok {
		return inf.precedence
	}
	return 0
}

func getPrefixName(op string) string {
	if pre, ok := prefixes[op]; ok {
		if pre.name == "" {
			return op
		}
		return pre.name
	}
	return op
}

func getInfixName(op string) string {
	if inf, ok := infixes[op]; ok {
		if inf.name == "" {
			return op
		}
		return inf.name
	}
	return op
}

func (me *parser) infixOp() (string, *parseError) {
	op := me.token.is
	if op == ">" {
		if me.peek().is == ">" {
			if er := me.eat(">"); er != nil {
				return "", er
			}
			op = ">>"
			me.replace(">", op)
		}
	}
	return op, nil
}
