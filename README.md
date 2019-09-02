# hymn-lang
compiles readable C code

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
* stack variables

### todo
* slices
* scope
* references to primitives
* free heap space
* interfaces
* dictionaries
* threads / async await
