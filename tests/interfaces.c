
#include <inttypes.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

typedef struct foo foo;
typedef struct bar bar;
typedef struct get_and_set get_and_set;
typedef struct special special;

struct get_and_set {
    int (*get)(void *self);
    void (*set)(void *self, int value);
};

struct foo {
    get_and_set interface;
    int one;
};

struct bar {
    get_and_set interface;
    int two;
};

struct special {
    get_and_set *first;
    get_and_set *second;
};

int foo_get(foo *self) {
    return self->one;
}

void foo_set(foo *self, int value) {
    self->one = value;
}

int bar_get(void *_self) {
    bar *self = (bar *)_self;
    return self->two;
}

void bar_set(void *_self, int value) {
    bar *self = (bar *)_self;
    self->two = value;
}

int main() {

    foo *f = calloc(1, sizeof(foo));
    f->interface.get = (int (*)(void *self))(&foo_get);
    f->interface.set = (void (*)(void *self, int value))(&foo_set);

    bar *b = calloc(1, sizeof(bar));
    b->interface.get = &bar_get;
    b->interface.set = &bar_set;

    special *s = calloc(1, sizeof(special));
    s->first = &f->interface;
    s->second = &b->interface;

    s->first->set(f, 3);
    s->second->set(b, 4);

    printf("foo %d\n", f->one);
    printf("bar %d\n", b->two);

    return 0;
}
