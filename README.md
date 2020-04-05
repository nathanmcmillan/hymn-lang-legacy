# Hymn Programming Language

Hymn is a programming language designed to simplify writing complex software.

It focuses on simplicity and visually pleasing syntax without sacrificing critical features such as static typing.

Hymn compiles to readable C code, and can easily be included in existing projects.

```
class vec
    x int
    y int

enum result<a,b>
    ok(value a)
    error(message b)

def main   
    v = vec(2, 3)
    r = result<vec,string>.ok(v)
    match r
        ok(o): echo("vec =", o.value.x, o.value.y)
        error(e): echo("error =", e.message)
```

## Links 
- [Homepage](https://hymn-lang.org)
- [Read the book](https://hymn-lang.org/book/index.html)
- [Learn by example](https://hymn-lang.org/learn-by-example/index.html)

## Why Hymn?

- Maybe<> type and strict match statements prevent null pointer exceptions
- Safe union types. Unions are used through enum types and always require matching
- Predictable runtime, there is no garbage collection
- Compiles to C. Fully usable with existing or future C projects
- Small binary size

1. Why not Python?
   - Interpreted languages are too slow for many use-cases.
   - Dynamic types can make large programs difficult to reason with.
2. Why not Golang?
   - Lack of generics can make some programs otherwise tedious to code. 
   - Lack of tagged union types reduces expressiveness.
3. Why not Rust?
   - Often difficult to understand, with a high entry barrier to learning
   - Often too much for less critical software
4. Why not C?
   - Lack of conveniences such as generics and name-spaces 
   - Often too easy to make critical mistakes
5. Why not C++?
   - Slow compile times compared to C
   - Considered bloated with many legacy problems

Visit the [website](https://hymn-lang.org) to learn more!

## Development

### Completed
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

### Not Started
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
* Need matching C code
