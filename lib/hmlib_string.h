#ifndef HMLIB_STRING_H
#define HMLIB_STRING_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <stdarg.h>
#include <stdint.h>
#include <inttypes.h>

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
int hmlib_string_len_int(const hmlib_string s);
void hmlib_string_free(const hmlib_string s);

char *hmlib_concat(const char *a, const char *b);
char *hmlib_concat_list(const char **ls, const int size);
char *hmlib_concat_varg(const int size, ...);

char *hmlib_char_to_string(const char ch);
char *hmlib_int_to_string(const int number);
char *hmlib_int8_to_string(const int8_t number);
char *hmlib_int16_to_string(const int16_t number);
char *hmlib_int32_to_string(const int32_t number);
char *hmlib_int64_to_string(const int64_t number);

char *hmlib_uint_to_string(const unsigned int number);
char *hmlib_uint8_to_string(const uint8_t number);
char *hmlib_uint16_to_string(const uint16_t number);
char *hmlib_uint32_to_string(const uint32_t number);
char *hmlib_uint64_to_string(const uint64_t number);

char *hmlib_float_to_string(const float number);
char *hmlib_float32_to_string(const float number);
char *hmlib_float64_to_string(const double number);

int hmlib_string_to_int(const char *str);
int8_t hmlib_string_to_int8(const char *str);
int16_t hmlib_string_to_int16(const char *str);
int32_t hmlib_string_to_int32(const char *str);
int64_t hmlib_string_to_int64(const char *str);

unsigned int hmlib_string_to_uint(const char *str);
uint8_t hmlib_string_to_uint8(const char *str);
uint16_t hmlib_string_to_uint16(const char *str);
uint32_t hmlib_string_to_uint32(const char *str);
uint64_t hmlib_string_to_uint64(const char *str);

float hmlib_string_to_float(const char *str);
float hmlib_string_to_float32(const char *str);
double hmlib_string_to_float64(const char *str);

#endif
