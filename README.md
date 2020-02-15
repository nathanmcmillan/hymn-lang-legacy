# hymn-lang
Hymn is a programming language designed to make writing simple imperitive programs easy.
It compiles to efficient, readable C code.

```
class vec<t>
    data t

enum either<a,b>
    first(value a)
    second(value b, additional string)

function main
    v = vec(data:"hello world")
    echo("data :=", v.data)
    e = either<int,float>.first(66)
    match e
        first(f) => echo("first :=", f.value)
        second(s) => echo("second :=", s.value, s.additional)
```

Learn more at https://hymn-lang.org

### Features
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
* $HYMN_MODULES environment variable

### Timeline
* bootstrapping compiler from golang to hymn
* references to primitives
* borrow checker
* free heap space
* interfaces / contraints (compile time check whether a class implements a set of functions)
* threads / async await (split function in half for each await)
* generate makefiles
* macros / def
* better error output
* language server protocol
* optimize printf for multiple strings to avoid concatenation 

### Bugs
* variable scoping

### Testing
* need negative tests
* need matching C code
