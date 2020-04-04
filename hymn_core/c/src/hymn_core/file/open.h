#ifndef HM_HYMN_CORE_FILE_OPEN_H
#define HM_HYMN_CORE_FILE_OPEN_H

#include "hymn_core/file/file.h"
#include "hymn_core/os_error/os_error.h"

enum HmFileOpen {
    HmFileOpenOk,
    HmFileOpenError
};

typedef enum HmFileOpen HmFileOpen;
typedef struct HmFileUnionOpen HmFileUnionOpen;
struct HmFileUnionOpen {
    HmFileOpen type;
    union {
        HmFile *ok;
        HmOsError *error;
    };
};

#endif
