#ifndef HM_HYMN_CORE_FILE_H
#define HM_HYMN_CORE_FILE_H

#include "hmlib_string.h"

#include "hymn_core/os_error/os_error.h"

typedef struct HmFile HmFile;
struct HmFile {
    hmlib_string content;
};

#include "open.h"
#endif
