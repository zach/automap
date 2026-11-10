// stb_truetype's binding to the allocation contract, and the one translation
// unit that compiles stb_truetype's implementation. stb_truetype's
// allocation hooks are macros its implementation reads, so they are defined
// here, before it is included.
// SPDX-License-Identifier: MIT-0
#include "automap/alloc.h"

#define STBTT_malloc(size, userdata) ((void)(userdata), am::alloc::allocate("stb-truetype", size))
#define STBTT_free(block, userdata) ((void)(userdata), am::alloc::release("stb-truetype", block))
#define STB_TRUETYPE_IMPLEMENTATION
#include <stb_truetype.h>
