#include "hmlib_dict.h"

HmLibDict *hmlib_dict()
{
  const int capacity = 10;
  HmLibDict *dict = malloc(sizeof(HmLibDict));
  dict->capacity = capacity;
  dict->table = calloc(capacity, sizeof(HmLibDictEntry));
  return dict;
}

int hmlib_dict_int_hash(HmLibDict *self, int key)
{
  return key % self->capacity;
}

int hmlib_dict_string_hash(HmLibDict *self, char *key)
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

HmLibDictEntry *hmlib_dict_entry(int key, char *value)
{
  HmLibDictEntry *entry = malloc(sizeof(HmLibDictEntry));
  entry->key = key;
  size_t len = strlen(value) + 1;
  entry->value = malloc(len);
  memcpy(entry->value, value, len);
  return entry;
}

void hmlib_dict_set(HmLibDict *self, int key, char *value)
{
  int bin = hmlib_dict_int_hash(self, key);
  HmLibDictEntry *next = self->table[bin];

  HmLibDictEntry *last = NULL;
  while (next != NULL && key != next->key)
  {
    last = next;
    next = next->next;
  }

  if (next != NULL && key == next->key)
  {
    free(next->value);
    size_t len = strlen(value) + 1;
    next->value = malloc(len);
    memcpy(next->value, value, len);
  }
  else
  {
    HmLibDictEntry *entry = hmlib_dict_entry(key, value);
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

char *hmlib_dict_get(HmLibDict *self, int key)
{
  int bin = hmlib_dict_int_hash(self, key);
  HmLibDictEntry *entry = self->table[bin];
  if (entry == NULL || key != entry->key)
  {
    return NULL;
  }
  return entry->value;
}

bool hmlib_dict_has(HmLibDict *self, int key)
{
  int bin = hmlib_dict_int_hash(self, key);
  HmLibDictEntry *entry = self->table[bin];
  if (entry == NULL || key != entry->key)
  {
    return false;
  }
  return true;
}

void hmlib_dict_delete(HmLibDict *self, int key)
{
  self->size--;
}

void hmlib_dict_clear(HmLibDict *self)
{
  self->size = 0;
}
