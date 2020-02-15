# hymn-lang
Hymn is a programming language designed to make writing simple imperitive programs easy.
It compiles to efficient, readable C code.

```
class vec<g>
    x int
    y float
    z g

enum either<a,b>
    first(value a)
    second(value b, additional string)

function main   
    v = vec(x:12, y:23.34, g:"hello world")
    echo("x :=", v.x, "y :=", v.y, "z :=", v.z)
    e = either<int,float>.first(66)
    match e
        first(f) => echo("first :=", f.value)
        second(s) => echo("second :=", s.value, s.additional)
```

Learn more at https://hymn-lang.org

### Features
* Generics
* Goto and Labels
* Enums with Unions
* Classes
* Match statements
* Walrus operator
* Defining stack or heap variables
* Function pointers
* Slices and arrays
* Class functions with generics
* Automatic or manual default parameters using '_'
* Configurable environment variables using $HYMN_MODULES

### Timeline
* Bootstrapping compiler from golang to hymn
* References to primitives
* Borrow checker
* Free heap space
* Interfaces / contraints (compile time check whether a class implements a set of functions)
* Threads / async await (split function in half for each await)
* Generate makefiles
* Macros / def
* Better error output
* Language server protocol
* Optimize printf for multiple strings to avoid concatenation 

### Bugs
* Variable scoping

### Testing
* Need negative tests
* Need matching C code
