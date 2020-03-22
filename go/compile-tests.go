package main

import (
	"path/filepath"
	"strings"
)

func (me *program) generateUnitTestsC() string {

	mainc := filepath.Join(me.outsourcedir, "main.c")

	var code strings.Builder

	code.WriteString("#include <stdlib.h>\n\n")

	code.WriteString("#include \"hymn/test/test.h\"\n\n")
	code.WriteString("#include \"test_boot/test_tokenizer.h\"\n\n")

	code.WriteString("int main() {\n")
	code.WriteString("    HmTest *const t = calloc(1, sizeof(HmTest));\n")
	code.WriteString("    hm_test_tokenizer_test_tokenizer(t);\n")
	code.WriteString("    return 0;\n")
	code.WriteString("}\n")

	write(mainc, code.String())

	return mainc
}
