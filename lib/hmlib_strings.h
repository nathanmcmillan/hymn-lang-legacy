#ifndef HMLIB_STRINGS_H
#define HMLIB_STRINGS_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <stdarg.h>

struct __attribute__((__packed__)) hmlib_string_head
{
    size_t len;
    size_t cap;
    char **chars;
};

typedef char *hmlib_string;
typedef struct hmlib_string_head hmlib_string_head;

hmlib_string hmlib_string_init_with_length(const char *init, size_t len);
hmlib_string hmlib_string_init(const char *init);
hmlib_string hmlib_string_concat(const hmlib_string a, const hmlib_string b);
size_t hmlib_string_len(const hmlib_string s);
void hmlib_string_free(const hmlib_string s);

char *hmlib_concat(const char *a, const char *b);
char *hmlib_concat_list(const char **ls, const int size);
char *hmlib_concat_varg(const int size, ...);
char *hmlib_int_to_string(const int number);
char *hmlib_float_to_string(const float number);
int hmlib_string_to_int(const char *str);
float hmlib_string_to_float(const char *str);

#endif
