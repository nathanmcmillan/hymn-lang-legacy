#include "029-arrays.h"

int *hm_029_arrays_foo(int *const z) {
    hmlib_string string_temp0 = hmlib_string_init("size: ");
    hmlib_string string_temp1 = hmlib_int_to_string(hmlib_slice_len(z));
    hmlib_string concat_temp0 = hmlib_concat(string_temp0, string_temp1);
    hmlib_string_free(string_temp0);
    hmlib_string_free(string_temp1);
    printf("%s\n", concat_temp0);
    hmlib_string_free(concat_temp0);
    int i;
    for (i = 0; i < hmlib_slice_len(z); i += 1) {
        hmlib_string string_temp2 = hmlib_string_init("z[");
        hmlib_string string_temp3 = hmlib_int_to_string(i);
        hmlib_string string_temp4 = hmlib_string_init("]: ");
        hmlib_string string_temp5 = hmlib_int_to_string(z[i]);
        hmlib_string concat_temp1 = hmlib_concat_varg(4, string_temp2, string_temp3, string_temp4, string_temp5);
        hmlib_string_free(string_temp2);
        hmlib_string_free(string_temp3);
        hmlib_string_free(string_temp4);
        hmlib_string_free(string_temp5);
        printf("%s\n", concat_temp1);
        hmlib_string_free(concat_temp1);
    }
    z[2] = 777;
    hmlib_string string_temp8 = hmlib_string_init("z[2]: ");
    hmlib_string string_temp9 = hmlib_int_to_string(z[2]);
    hmlib_string concat_temp3 = hmlib_concat(string_temp8, string_temp9);
    hmlib_string_free(string_temp8);
    hmlib_string_free(string_temp9);
    printf("%s\n", concat_temp3);
    hmlib_string_free(concat_temp3);
    return z;
}

int main() {
    int *a = malloc(3 * sizeof(int));
    hmlib_string string_temp0 = hmlib_string_init("size: ");
    hmlib_string string_temp1 = hmlib_int_to_string(3);
    hmlib_string concat_temp0 = hmlib_concat(string_temp0, string_temp1);
    hmlib_string_free(string_temp0);
    hmlib_string_free(string_temp1);
    printf("%s\n", concat_temp0);
    hmlib_string_free(concat_temp0);
    int i;
    for (i = 0; i < 3; i += 1) {
        a[i] = 10 + i;
    }
    for (i = 0; i < 3; i += 1) {
        hmlib_string string_temp2 = hmlib_string_init("a[");
        hmlib_string string_temp3 = hmlib_int_to_string(i);
        hmlib_string string_temp4 = hmlib_string_init("]: ");
        hmlib_string string_temp5 = hmlib_int_to_string(a[i]);
        hmlib_string concat_temp1 = hmlib_concat_varg(4, string_temp2, string_temp3, string_temp4, string_temp5);
        hmlib_string_free(string_temp2);
        hmlib_string_free(string_temp3);
        hmlib_string_free(string_temp4);
        hmlib_string_free(string_temp5);
        printf("%s\n", concat_temp1);
        hmlib_string_free(concat_temp1);
    }
    hmlib_slice slice_temp0 = hmlib_array_to_slice(a, sizeof(int), 3);
    hm_029_arrays_foo(slice_temp0);
    hmlib_slice_free(slice_temp0);
    hmlib_string string_temp6 = hmlib_string_init("a[2]: ");
    hmlib_string string_temp7 = hmlib_int_to_string(a[2]);
    hmlib_string concat_temp2 = hmlib_concat(string_temp6, string_temp7);
    hmlib_string_free(string_temp6);
    hmlib_string_free(string_temp7);
    printf("%s\n", concat_temp2);
    hmlib_string_free(concat_temp2);
    free(a);
    return 0;
}
