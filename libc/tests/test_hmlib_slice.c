#include "../lib/hmlib_slice.h"

struct vec {
    float x;
    float y;
    float z;
};

typedef struct vec vec;

int main() {
    vec **a = hmlib_slice_init(sizeof(vec), 1, 1);
    printf("addr %lld\n", (unsigned long long)a);
    printf("length = %ld\n", hmlib_slice_len(a));

    a[0] = malloc(sizeof(vec));
    a[0]->x = 44;
    printf("get 1 = %f\n", a[0]->x);

    vec *w = malloc(sizeof(vec));
    w->x = 66;
    printf("w = %f\n", w->x);

    a = hmlib_slice_push(a, w);
    printf("new length = %ld\n", hmlib_slice_len(a));
    printf("get 2 = %f\n", a[0]->x);
    printf("get 3 = %f\n", a[1]->x);

    vec **b = hmlib_slice_init(sizeof(vec), 4, 4);
    b[0] = malloc(sizeof(vec));
    b[0]->x = 88;

    a = hmlib_slice_expand(a, b);
    printf("new length = %ld\n", hmlib_slice_len(a));
    printf("get 4 = %f\n", a[0]->x);
    printf("get 5 = %f\n", a[1]->x);
    printf("get 6 = %f\n", a[2]->x);

    hmlib_slice_free(a);

    return 0;
}
