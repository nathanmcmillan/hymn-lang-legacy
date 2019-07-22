#ifndef HMLIB_DICT_H
#define HMLIB_DICT_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <limits.h>

typedef struct HmLibDictEntry HmLibDictEntry;
typedef struct HmLibDict HmLibDict;

struct HmLibDictEntry
{
    int key;
    char *value;
    HmLibDictEntry *next;
};

struct HmLibDict
{
    int size;
    int capacity;
    HmLibDictEntry **table;
};

HmLibDict *hmlib_dict();
int hmlib_dict_string_hash(HmLibDict *self, char *key);
HmLibDictEntry *hmlib_dict_entry(int key, char *value);
void hmlib_dict_set(HmLibDict *self, int key, char *value);
char *hmlib_dict_get(HmLibDict *self, int key);
bool hmlib_dict_has(HmLibDict *self, int key);
void hmlib_dict_delete(HmLibDict *self, int key);
void hmlib_dict_clear(HmLibDict *self);

#endif
