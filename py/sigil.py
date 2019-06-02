import sys
import tokens as tokenize
import parse as parser
import json
import ccode
import os


def main():
    tokens = None
    try:
        a = sys.argv[1]
        with open(a) as f:
            source = f.read()
        print("=== source ===")
        print(source)
        print("=== tokens ===")
        tokens = tokenize.read(source)
        print("=== parse ===")
        program = parser.read(tokens)
        with open("out/ast.json", "w") as f:
            json.dump(program, f, indent=2, sort_keys=True)
        print("=== c-code ===")
        cfile = "out/main.c"
        ccode.main(cfile, program, 0)
        print("=== gcc ===")
        app = "out/main.app"
        if os.path.isfile(app):
            os.remove(app)
        os.system("gcc " + cfile + " -o " + app)
        if os.path.isfile(app):
            print("=== run ===")
            os.system(app)
        print("===")
    except AssertionError as err:
        print("=== error ===")
        print(err)
        if tokens:
            i = 0
            for t in tokens:
                print(str(i) + ":", t)
                i += 1


if __name__ == "__main__":
    main()
