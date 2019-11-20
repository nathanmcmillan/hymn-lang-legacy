# hymn-lang
Hymn is a programming language designed to make writing simple imperitive programs easy.
It compiles to efficient, readable C code.

```
type foo<t>
  data t

main
  f = foo(data:"hello world")
  echo(f.data)
```

Learn more at https://hymn-lang.org

### features
* generics
* goto + labels
* enums + unions
* structs
* matching
* walrus operator
* stack variables
* function pointers
* slices and arrays
* class functions with generics
* "_" for default parameters during class allocation

### timeline
* hash maps
* file input / output
* bootstrapping compiler from golang to hymn
* JSON format tokens and parse tree
* references to primitives
* borrow checker
* free heap space
* interfaces / contraints (compile time check whether a class implements a set of functions)
* threads / async await (split function in half for each await)
* macros / def
* better error output
* language server protocol
* optimize printf for multiple strings to avoid concatenation 

### bugs
* variable scoping

### testing
* need negative tests
* need regex matching C code

### ideas
* "transfer" keyword replaces "return" when giving a strong pointer is desired
