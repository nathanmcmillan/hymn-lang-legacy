#ifndef HMLIB_STRINGS_H
#define HMLIB_STRINGS_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <stdarg.h>

char *hmlib_concat(const char *a, const char *b);
char *hmlib_concat_list(const char **ls, const int size);
char *hmlib_concat_varg(const int size, ...);
char *hmlib_int_to_string(const int number);
char *hmlib_float_to_string(const float number);
int hmlib_string_to_int(const char *str);
float hmlib_string_to_float(const char *str);

#endif
