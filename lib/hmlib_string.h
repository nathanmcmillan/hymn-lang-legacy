#ifndef HMLIB_STRING_H
#define HMLIB_STRING_H

#include <inttypes.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef char *hmlib_string;
typedef struct hmlib_string_head hmlib_string_head;

struct __attribute__((__packed__)) hmlib_string_head
{
    size_t length;
    size_t capacity;
    char **chars;
};

hmlib_string hmlib_string_init_with_length(const char *init, size_t length);
hmlib_string hmlib_string_init(const char *init);

size_t hmlib_string_len_size(const hmlib_string s);
int hmlib_string_len(const hmlib_string s);
size_t hmlib_string_cap_size(const hmlib_string s);
int hmlib_string_cap(const hmlib_string s);
void hmlib_string_free(const hmlib_string s);

hmlib_string hmlib_concat(const hmlib_string a, const hmlib_string b);
hmlib_string hmlib_concat_list(const hmlib_string *list, const int size);
hmlib_string hmlib_concat_varg(const int size, ...);

hmlib_string hmlib_string_append(const hmlib_string a, const char *b);

int hmlib_string_compare(const hmlib_string a, const hmlib_string b);
bool hmlib_string_equal(const hmlib_string a, const hmlib_string b);

hmlib_string hmlib_char_to_string(const char ch);
hmlib_string hmlib_int_to_string(const int number);
hmlib_string hmlib_int8_to_string(const int8_t number);
hmlib_string hmlib_int16_to_string(const int16_t number);
hmlib_string hmlib_int32_to_string(const int32_t number);
hmlib_string hmlib_int64_to_string(const int64_t number);

hmlib_string hmlib_uint_to_string(const unsigned int number);
hmlib_string hmlib_uint8_to_string(const uint8_t number);
hmlib_string hmlib_uint16_to_string(const uint16_t number);
hmlib_string hmlib_uint32_to_string(const uint32_t number);
hmlib_string hmlib_uint64_to_string(const uint64_t number);

hmlib_string hmlib_float_to_string(const float number);
hmlib_string hmlib_float32_to_string(const float number);
hmlib_string hmlib_float64_to_string(const double number);

int hmlib_string_to_int(const hmlib_string str);
int8_t hmlib_string_to_int8(const hmlib_string str);
int16_t hmlib_string_to_int16(const hmlib_string str);
int32_t hmlib_string_to_int32(const hmlib_string str);
int64_t hmlib_string_to_int64(const hmlib_string str);

unsigned int hmlib_string_to_uint(const hmlib_string str);
uint8_t hmlib_string_to_uint8(const hmlib_string str);
uint16_t hmlib_string_to_uint16(const hmlib_string str);
uint32_t hmlib_string_to_uint32(const hmlib_string str);
uint64_t hmlib_string_to_uint64(const hmlib_string str);

float hmlib_string_to_float(const hmlib_string str);
float hmlib_string_to_float32(const hmlib_string str);
double hmlib_string_to_float64(const hmlib_string str);

#endif
