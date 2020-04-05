#ifndef HMLIB_FILES_H
#define HMLIB_FILES_H

#include <stdarg.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "hmlib_mem.h"
#include "hmlib_string.h"

hmlib_string hmlib_cat(const char *path);
void hmlib_write(const char *path, const char *content);

#endif
