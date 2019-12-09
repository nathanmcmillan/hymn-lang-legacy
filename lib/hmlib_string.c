#include "hmlib_string.h"

hmlib_string_head *hmlib_string_head_init(const size_t length, const size_t capacity) {
    size_t memory = sizeof(hmlib_string_head) + length + 1;
    hmlib_string_head *head = (hmlib_string_head *)malloc(memory);
    memset(head, 0, memory);
    head->length = length;
    head->capacity = capacity;
    return head;
}

hmlib_string hmlib_string_init_with_length(const char *init, const size_t length) {
    hmlib_string_head *head = hmlib_string_head_init(length, length);
    char *s = (char *)(head + 1);
    memcpy(s, init, length);
    s[length] = '\0';
    return (hmlib_string)s;
}

hmlib_string hmlib_string_init(const char *init) {
    size_t len = strlen(init);
    return hmlib_string_init_with_length(init, len);
}

size_t hmlib_string_len_size(const hmlib_string s) {
    hmlib_string_head *head = (hmlib_string_head *)((char *)s - sizeof(hmlib_string_head));
    return head->length;
}

int hmlib_string_len(const hmlib_string s) {
    return (int)hmlib_string_len_size(s);
}

size_t hmlib_string_cap_size(const hmlib_string s) {
    hmlib_string_head *head = (hmlib_string_head *)((char *)s - sizeof(hmlib_string_head));
    return head->capacity;
}

int hmlib_string_cap(const hmlib_string s) {
    return (int)hmlib_string_cap_size(s);
}

void hmlib_string_free(const hmlib_string s) {
    free((char *)s - sizeof(hmlib_string_head));
}

hmlib_string hmlib_concat(const hmlib_string a, const hmlib_string b) {
    const size_t len1 = hmlib_string_len_size(a);
    const size_t len2 = hmlib_string_len_size(b);
    const size_t len = len1 + len2;
    hmlib_string_head *head = hmlib_string_head_init(len, len);
    char *s = (char *)(head + 1);
    memcpy(s, a, len1);
    memcpy(s + len1, b, len2 + 1);
    s[len] = '\0';
    return (hmlib_string)s;
}

hmlib_string hmlib_concat_list(const hmlib_string *list, const int size) {
    size_t len = 0;
    for (int i = 0; i < size; i++) {
        len += hmlib_string_len_size(list[i]);
    }
    hmlib_string_head *head = hmlib_string_head_init(len, len);
    char *s = (char *)(head + 1);
    size_t pos = 0;
    for (int i = 0; i < size; i++) {
        size_t len_i = hmlib_string_len_size(list[i]);
        memcpy(s + pos, list[i], len_i);
        pos += len_i;
    }
    s[len] = '\0';
    return (hmlib_string)s;
}

hmlib_string hmlib_concat_varg(const int size, ...) {
    va_list ap;

    size_t len = 0;
    va_start(ap, size);
    for (int i = 0; i < size; i++) {
        len += hmlib_string_len_size(va_arg(ap, hmlib_string));
    }
    va_end(ap);

    hmlib_string_head *head = hmlib_string_head_init(len, len);
    char *s = (char *)(head + 1);

    size_t pos = 0;
    va_start(ap, size);
    for (int i = 0; i < size; i++) {
        const hmlib_string param = va_arg(ap, hmlib_string);
        size_t len_i = hmlib_string_len_size(param);
        memcpy(s + pos, param, len_i);
        pos += len_i;
    }
    va_end(ap);

    s[len] = '\0';
    return (hmlib_string)s;
}

hmlib_string hmlib_substring(const hmlib_string a, const size_t start, const size_t end) {
    const size_t len = end - start;
    hmlib_string_head *head = hmlib_string_head_init(len, len);
    char *s = (char *)(head + 1);
    memcpy(s, a + start, len);
    s[len] = '\0';
    return (hmlib_string)s;
}

hmlib_string hmlib_string_append(const hmlib_string a, const char *b) {
    const size_t len1 = hmlib_string_len_size(a);
    const size_t len2 = strlen(b);
    const size_t len = len1 + len2;
    hmlib_string_head *head = hmlib_string_head_init(len, len);
    char *s = (char *)(head + 1);
    memcpy(s, a, len1);
    memcpy(s + len1, b, len2 + 1);
    s[len] = '\0';
    return (hmlib_string)s;
}

int hmlib_string_compare(const hmlib_string a, const hmlib_string b) {
    return strcmp(a, b);
}

bool hmlib_string_equal(const hmlib_string a, const hmlib_string b) {
    int comparison = hmlib_string_compare(a, b);
    return comparison == 0;
}

hmlib_string hmlib_char_to_string(const char ch) {
    char *str = malloc(2);
    str[0] = ch;
    str[1] = '\0';
    hmlib_string s = hmlib_string_init_with_length(str, 1);
    free(str);
    return s;
}

hmlib_string hmlib_int_to_string(const int number) {
    int len = snprintf(NULL, 0, "%d", number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%d", number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_int8_to_string(const int8_t number) {
    int len = snprintf(NULL, 0, "%" PRId8, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId8, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_int16_to_string(const int16_t number) {
    int len = snprintf(NULL, 0, "%" PRId16, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId16, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_int32_to_string(const int32_t number) {
    int len = snprintf(NULL, 0, "%" PRId32, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId32, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_int64_to_string(const int64_t number) {
    int len = snprintf(NULL, 0, "%" PRId64, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId64, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_uint_to_string(const unsigned int number) {
    int len = snprintf(NULL, 0, "%u", number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%u", number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_uint8_to_string(const uint8_t number) {
    int len = snprintf(NULL, 0, "%" PRId8, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId8, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_uint16_to_string(const uint16_t number) {
    int len = snprintf(NULL, 0, "%" PRId16, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId16, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_uint32_to_string(const uint32_t number) {
    int len = snprintf(NULL, 0, "%" PRId32, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId32, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_uint64_to_string(const uint64_t number) {
    int len = snprintf(NULL, 0, "%" PRId64, number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%" PRId64, number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_float_to_string(const float number) {
    int len = snprintf(NULL, 0, "%f", number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%f", number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

hmlib_string hmlib_float32_to_string(const float number) {
    return hmlib_float_to_string(number);
}

hmlib_string hmlib_float64_to_string(const double number) {
    int len = snprintf(NULL, 0, "%f", number);
    char *str = malloc(len + 1);
    snprintf(str, len + 1, "%f", number);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}

int hmlib_string_to_int(const hmlib_string str) {
    return (int)strtol(str, NULL, 10);
}

int8_t hmlib_string_to_int8(const hmlib_string str) {
    return (int8_t)strtol(str, NULL, 10);
}

int16_t hmlib_string_to_int16(const hmlib_string str) {
    return (int16_t)strtol(str, NULL, 10);
}

int32_t hmlib_string_to_int32(const hmlib_string str) {
    return (int32_t)strtol(str, NULL, 10);
}

int64_t hmlib_string_to_int64(const hmlib_string str) {
    return (int64_t)strtoll(str, NULL, 10);
}

unsigned int hmlib_string_to_uint(const hmlib_string str) {
    return (unsigned int)strtoul(str, NULL, 10);
}

uint8_t hmlib_string_to_uint8(const hmlib_string str) {
    return (uint8_t)strtoul(str, NULL, 10);
}

uint16_t hmlib_string_to_uint16(const hmlib_string str) {
    return (uint16_t)strtoul(str, NULL, 10);
}

uint32_t hmlib_string_to_uint32(const hmlib_string str) {
    return (uint32_t)strtoul(str, NULL, 10);
}

uint64_t hmlib_string_to_uint64(const hmlib_string str) {
    return (uint64_t)strtoull(str, NULL, 10);
}

float hmlib_string_to_float(const hmlib_string str) {
    return strtof(str, NULL);
}

float hmlib_string_to_float32(const hmlib_string str) {
    return hmlib_string_to_float(str);
}

double hmlib_string_to_float64(const hmlib_string str) {
    return strtod(str, NULL);
}

hmlib_string hmlib_format(const hmlib_string f, ...) {
    va_list ap;
    va_start(ap, f);
    int len = vsnprintf(NULL, 0, f, ap);
    va_end(ap);
    char *str = malloc((len + 1) * sizeof(char));
    va_start(ap, f);
    len = vsnprintf(str, len + 1, f, ap);
    va_end(ap);
    hmlib_string s = hmlib_string_init_with_length(str, len);
    free(str);
    return s;
}
