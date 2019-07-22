#ifndef HMLIB_DICT_STR_STR_H
#define HMLIB_DICT_STR_STR_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <limits.h>

typedef struct HmLibDictStrStrEntry HmLibDictStrStrEntry;
typedef struct HmLibDictStrStr HmLibDictStrStr;

struct HmLibDictStrStrEntry
{
    char *key;
    char *value;
    HmLibDictStrStrEntry *next;
};

struct HmLibDictStrStr
{
    int size;
    int capacity;
    HmLibDictStrStrEntry **table;
};

HmLibDictStrStr *hmlib_dict_str_str();
int hmlib_dict_string_hash(HmLibDictStrStr *self, char *key);
HmLibDictStrStrEntry *hmlib_dict_str_str_entry(char *key, char *value);
void hmlib_dict_str_str_set(HmLibDictStrStr *self, char *key, char *value);
char *hmlib_dict_str_str_get(HmLibDictStrStr *self, char *key);
bool hmlib_dict_str_str_has(HmLibDictStrStr *self, char *key);
void hmlib_dict_str_str_delete(HmLibDictStrStr *self, char *key);
void hmlib_dict_str_str_clear(HmLibDictStrStr *self);

#endif
