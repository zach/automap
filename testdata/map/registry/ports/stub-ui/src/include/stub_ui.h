/* stub-ui: a tiny immediate-mode UI library for Automap's fixture map.
 * Consumers must define STUB_UI_WIDE_IDS=1 to match how the library is
 * built. MIT-0. */
#ifndef STUB_UI_H
#define STUB_UI_H

#include <stdbool.h>
#include <stddef.h>

#if defined(STUB_UI_WIDE_IDS) && STUB_UI_WIDE_IDS
#define STUB_UI_ID_BITS 64
#else
#define STUB_UI_ID_BITS 32
#endif

/* MSVC's linker refuses to link translation units that disagree. */
#ifdef _MSC_VER
#if STUB_UI_ID_BITS == 64
#pragma detect_mismatch("STUB_UI_WIDE_IDS", "1")
#else
#pragma detect_mismatch("STUB_UI_WIDE_IDS", "0")
#endif
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef struct stub_ui_allocator {
    void* (*alloc)(size_t size, void* user);
    void (*release)(void* block, void* user);
    void* user;
} stub_ui_allocator;

typedef bool (*stub_ui_assert_fn)(const char* expression, const char* file, int line);

/* Allocation hooks; install before stub_ui_create_context. */
void stub_ui_set_allocator(const stub_ui_allocator* allocator);

/* Stores a copy of name, allocated with the current allocator. */
void stub_ui_set_name(const char* name);

void stub_ui_set_assert_handler(stub_ui_assert_fn handler);
void stub_ui_fail_test_assertion(void);

bool stub_ui_create_context(void);
void stub_ui_destroy_context(void);

bool stub_ui_attach(void* native_window);
void stub_ui_detach(void);
void stub_ui_enable_input(void);
void stub_ui_disable_input(void);

void stub_ui_feed_event(long long id, int x, int y);
void stub_ui_new_frame(void);
long long stub_ui_last_event_id(void);

/* Whether the caller's STUB_UI_ID_BITS matches the library's. */
bool stub_ui_abi_check(int id_bits);

#ifdef __cplusplus
}
#endif

#endif
