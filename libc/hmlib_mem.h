#ifndef HMLIB_MEM_H

#include <stdio.h>
#include <stdlib.h>

void *hmlib_malloc(size_t size);
void *hmlib_realloc(void *mem, size_t size);

#endif
