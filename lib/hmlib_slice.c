#include "hmlib_slice.h"

hmlib_slice_head *hmlib_slice_head_init(const size_t length)
{
  size_t memory = sizeof(hmlib_slice_head) + length * sizeof(void *);
  hmlib_slice_head *head = malloc(memory);
  memset(head, 0, memory);
  head->length = length;
  head->capacity = length;
  return head;
}

hmlib_slice_head *hmlib_slice_get_head(const hmlib_slice a)
{
  return a - sizeof(hmlib_slice_head);
}

hmlib_slice hmlib_slice_init(const size_t length)
{
  hmlib_slice_head *head = hmlib_slice_head_init(length);
  return (void *)head + sizeof(hmlib_slice_head);
}

void hmlib_slice_free(const hmlib_slice a)
{
  hmlib_slice_head *head = hmlib_slice_get_head(a);
  free(head);
}

size_t hmlib_slice_len(const hmlib_slice a)
{
  hmlib_slice_head *head = hmlib_slice_get_head(a);
  return head->length;
}

int hmlib_slice_len_int(const hmlib_slice a)
{
  return (int)hmlib_slice_len(a);
}

hmlib_slice hmlib_slice_expand(const hmlib_slice a, const hmlib_slice b)
{
  hmlib_slice_head *head_a = hmlib_slice_get_head(a);
  hmlib_slice_head *head_b = hmlib_slice_get_head(b);
  size_t length_a = head_a->length;
  size_t length_b = head_b->length;
  size_t length = length_a + length_b;
  size_t memory = sizeof(hmlib_slice_head) + length * sizeof(void *);
  hmlib_slice_head *head = realloc(head_a, memory);
  memcpy((void *)head + sizeof(hmlib_slice_head) + length_a * sizeof(void *), b, length_b * sizeof(void *));
  head->length = length;
  head->capacity = length;
  return (void *)head + sizeof(hmlib_slice_head);
}

hmlib_slice hmlib_slice_push(const hmlib_slice a, void *const b)
{
  hmlib_slice_head *head_a = hmlib_slice_get_head(a);
  size_t length = head_a->length + 1;
  size_t memory = sizeof(hmlib_slice_head) + length * sizeof(void *);
  hmlib_slice_head *head = realloc(head_a, memory);
  head->length = length;
  head->capacity = length;
  hmlib_slice data = (void *)head + sizeof(hmlib_slice_head);
  ((void **)data)[length - 1] = b;
  return data;
}
