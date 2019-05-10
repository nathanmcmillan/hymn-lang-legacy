import sys
import tokens as tokenize
import parse as parser
import compile as compiler
import ccode
import os


def main():
    a = sys.argv[1]
    with open(a) as f:
        source = f.read()
    print("=== source ===")
    print(source)
    print("=== tokens ===")
    tokens = tokenize.read(source)
    print("=== parse ===")
    program = parser.read(tokens)
    print("=== compile ===")
    compiler.read(program)
    print("=== c-code ===")
    with open("c-code/ss.c", "w") as f:
        ccode.write(f, program, 0)
    print("=== gcc ===")
    os.system("gcc c-code/ss.c -o c-code/ss.app")
    print("=== run ===")
    os.system("c-code/ss.app")
    print("===")


if __name__ == "__main__":
    try:
        main()
    except AssertionError as err:
        print(err)
