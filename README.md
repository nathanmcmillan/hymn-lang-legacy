# hymn-lang
compiles to readable C code

```
type foo<T>
  data T

main
  f = foo<string>
    data: "hello world"
  echo(f.data)
```

### features
* generics
* goto
* labels
* enums
* unions
* structs
* matching
* walrus

### todo
* slices
* scope
* stack variables
* references to primitives
* free heap space
* interfaces
* dictionaries
