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

### timeline
* class functions with generics
* hash maps
* file input / output
* bootstrapping compiler from golang to hymn
* JSON format tokens and parse tree
* correct scoping for functions
* references to primitives
* borrow checker
* free heap space
* interfaces (maybe?)
* threads / async await
* macros / def
* better error output
* language server protocol
* use "_" for default parameters during class allocation

### testing
* need negative tests
* need regex matching C code

### ideas
* "transfer" keyword replaces "return" when giving a strong pointer is desired
