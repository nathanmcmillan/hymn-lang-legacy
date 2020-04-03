package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (me *program) generateUnitTestsC() string {

	main := filepath.Join(me.outsourcedir, "main.c")

	var code strings.Builder
	var include strings.Builder

	include.WriteString("#include <stdlib.h>\n\n")
	include.WriteString("#include <time.h>\n\n")

	include.WriteString("#include \"hymn/test/test.h\"\n\n")

	code.WriteString("int main() {\n")

	code.WriteString("    time_t now = time(NULL);\n")
	code.WriteString("    struct tm *timeinfo = localtime(&now);\n")
	code.WriteString("    printf(\"%s\\n\", asctime(timeinfo));\n")

	code.WriteString("    HmTestStats *stats = hm_test_start();\n")

	for _, m := range me.modules {
		if !strings.HasPrefix(m.name, "test_") {
			continue
		}
		requires := fmt.Sprintf("#include \"%s/%s/%s.h\"\n\n", m.pack[0], m.name, m.name)
		include.WriteString(requires)
		for _, fn := range m.functionOrder {
			f := m.functions[fn]
			if strings.HasPrefix(f.getname(), "test_") {
				expr := fmt.Sprintf("    hm_test_run_test(\"%s\", \"%s\", \"%s\", %s, stats);\n", m.pack[0], m.name, f.getname(), f.getcname())
				code.WriteString(expr)
			}
		}
	}

	code.WriteString("    return hm_test_end(stats);\n")
	code.WriteString("}\n")

	out := include.String() + code.String()

	write(main, out)

	return main
}
