package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (me *program) generateUnitTestsC() string {

	main := filepath.Join(me.outsourcedir, "main.c")

	var code strings.Builder

	code.WriteString("#include <stdlib.h>\n\n")
	code.WriteString("#include <time.h>\n\n")

	code.WriteString("#include \"hymn/test/test.h\"\n\n")
	code.WriteString("#include \"test_boot/test_tokenizer/test_tokenizer.h\"\n\n")

	code.WriteString("int main() {\n")

	code.WriteString("    time_t rawtime;\n")
	code.WriteString("    struct tm *timeinfo;\n")
	code.WriteString("    time(&rawtime);\n")
	code.WriteString("    timeinfo = localtime(&rawtime);\n")
	code.WriteString("    printf(\"%s\\n\", asctime(timeinfo));\n")

	code.WriteString("    HmTestStats *stats = hm_test_start();\n")

	for _, m := range me.modules {
		if !strings.HasPrefix(m.name, "test_") {
			continue
		}
		for _, fn := range m.functionOrder {
			f := m.functions[fn]
			if strings.HasPrefix(f.getname(), "test_") {
				expr := fmt.Sprintf("    hm_test_run_test(\"%s\", \"%s\", \"%s\", %s, stats);\n", m.pack[0], m.name, f.getname(), f.getcname())
				code.WriteString(expr)
			}
		}
	}

	code.WriteString("    time(&rawtime);\n")
	code.WriteString("    timeinfo = localtime(&rawtime);\n")
	code.WriteString("    printf(\"\\n%s\", asctime(timeinfo));\n")

	code.WriteString("    return hm_test_end(stats);\n")
	code.WriteString("}\n")

	write(main, code.String())

	return main
}
