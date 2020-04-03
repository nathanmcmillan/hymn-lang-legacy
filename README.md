# hymn-lang
Hymn is a programming language designed to make writing simple imperitive programs easy.
It compiles to efficient, readable C code.

```
class vec
    x int
    y int

enum result<a,b>
    ok(value a)
    error(value b)

def main   
    v = vec(2, 3)
    e = result<vec,string>.ok(v)
    match e
        ok(o): echo("vec =", o.value.x, o.value.y)
        error(e): echo("error =", e.value)
```

## Why

Hymn aims to make it as easy and safe as possible to compile binary programs.

1. Why not Python?
   - Slow
   - No types 
2. Why not Golang?
   - No generics
   - No enum types
   - Error handling
3. Why not Rust?
   - Difficult to understand
   - High entry barrier
   - Overkill for less critical software
4. Why not C?
   - Easy to make mistakes
   - No generics
   - No Namespaces
5. Why not C++?
   - Bloated
   - Slow compile times
   - Legacy cruft
   - Hard to read

## Links 
[Homepage](https://hymn-lang.org)
[Read the book](https://hymn-lang.org/site/book/index.html)
[Learn by example](https://hymn-lang.org/site/learn-by-example/index.html)

## Development

### Features
* Generics
* Goto and Labels
* Enums with Unions
* Classes
* Match statements
* Defining stack or heap variables
* Function pointers
* Slices and arrays
* Class functions with generics
* Automatic or manual default parameters using `_`
* `$HYMN_LIBC` environment variable and -d flag locates the standard hymn c library
* Package management using `$HYMN_PACKAGES` environment variable
* Multiline string declaration using '\'
* Interfaces

### Timeline
* References to primitives
* Multiple return values
* Bootstrapping compiler from golang to hymn
* Borrow checker
* Free heap space
* Threads / async await (split function in half for each await)
* Generate makefiles
* Macros / def
* Better error output
* Language server protocol: The compiler should have flags for how to format/output found problems and at what point to stop
* Optimize printf for multiple strings to avoid concatenation

### Libraries
The standard libraries will need to include the following
* ref/ptr class for holding pointers to primitives
* tuple class for returning multiple values
* hashmap, hashset, list classes for standard data structures
* either enum for union returns
* string builder class

### Bugs
* Variable scoping

### Testing
* Need negative tests
* Need matching C code
