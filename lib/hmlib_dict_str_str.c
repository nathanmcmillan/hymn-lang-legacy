#include "hmlib_dict_str_str.h"

HmLibDictStrStr *hmlib_dict_str_str()
{
  const int capacity = 10;
  HmLibDictStrStr *dict = malloc(sizeof(HmLibDictStrStr));
  dict->capacity = capacity;
  dict->table = calloc(capacity, sizeof(HmLibDictStrStrEntry));
  return dict;
}

int hmlib_dict_string_hash(HmLibDictStrStr *self, char *key)
{
  int index = 0;
  unsigned long int value = 0;
  size_t len = strlen(key);
  while (value < ULONG_MAX && index < len)
  {
    value = (value << 8) + key[index];
    index++;
  }
  return value % self->capacity;
}

HmLibDictStrStrEntry *hmlib_dict_str_str_entry(char *key, char *value)
{
  HmLibDictStrStrEntry *entry = malloc(sizeof(HmLibDictStrStrEntry));
  entry->key = key;
  size_t len = strlen(value) + 1;
  entry->value = malloc(len);
  memcpy(entry->value, value, len);
  return entry;
}

void hmlib_dict_str_str_set(HmLibDictStrStr *self, char *key, char *value)
{
  int bin = hmlib_dict_string_hash(self, key);
  HmLibDictStrStrEntry *next = self->table[bin];

  HmLibDictStrStrEntry *last = NULL;
  while (next != NULL && strcmp(key, next->key) > 0)
  {
    last = next;
    next = next->next;
  }

  if (next != NULL && strcmp(key, next->key) > 0)
  {
    free(next->value);
    size_t len = strlen(value) + 1;
    next->value = malloc(len);
    memcpy(next->value, value, len);
  }
  else
  {
    HmLibDictStrStrEntry *entry = hmlib_dict_str_str_entry(key, value);
    if (next == self->table[bin])
    {
      entry->next = next;
      self->table[bin] = entry;
    }
    else if (next == NULL)
    {
      last->next = entry;
    }
    else
    {
      entry->next = next;
      last->next = entry;
    }
    self->size++;
  }
}

char *hmlib_dict_str_str_get(HmLibDictStrStr *self, char *key)
{
  int bin = hmlib_dict_string_hash(self, key);
  HmLibDictStrStrEntry *entry = self->table[bin];
  if (entry == NULL || strcmp(key, entry->key) != 0)
  {
    return NULL;
  }
  return entry->value;
}

bool hmlib_dict_str_str_has(HmLibDictStrStr *self, char *key)
{
  int bin = hmlib_dict_string_hash(self, key);
  HmLibDictStrStrEntry *entry = self->table[bin];
  if (entry == NULL || strcmp(key, entry->key) != 0)
  {
    return false;
  }
  return true;
}

void hmlib_dict_str_str_delete(HmLibDictStrStr *self, char *key)
{
  self->size--;
}

void hmlib_dict_str_str_clear(HmLibDictStrStr *self)
{
  self->size = 0;
}
