/* stub-window: a tiny windowing library for Automap's fixture map. It
 * "opens" windows only when STUB_WINDOW_HEADLESS=1 is set, as there is no
 * display. MIT-0. */
#ifndef STUB_WINDOW_H
#define STUB_WINDOW_H

#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef void* (*stub_window_alloc_fn)(size_t size);
typedef void (*stub_window_free_fn)(void* block);
typedef void (*stub_window_log_fn)(const char* message);

typedef struct stub_window stub_window;

#if defined(STUB_WINDOW_WIDE_EVENTS) && STUB_WINDOW_WIDE_EVENTS
typedef long long stub_window_event_id;
#else
typedef int stub_window_event_id;
#endif

typedef struct stub_window_event {
    stub_window_event_id id;
    int x;
    int y;
} stub_window_event;

/* Allocation hooks; install before any other call. */
void stub_window_set_allocator(stub_window_alloc_fn alloc, stub_window_free_fn release);
void stub_window_get_allocator(stub_window_alloc_fn* alloc, stub_window_free_fn* release);

/* Stores a copy of name, allocated with the current allocator. */
void stub_window_set_name(const char* name);

void stub_window_set_log(stub_window_log_fn log);
void stub_window_log_test_message(void);

bool stub_window_init(void);
stub_window* stub_window_open(const char* title, int width, int height);
void stub_window_close(stub_window* window);
void stub_window_shutdown(void);

void stub_window_push_event(stub_window_event event);
bool stub_window_poll(stub_window_event* event);

/* Whether the caller's view of the ABI matches the library's. */
bool stub_window_abi_check(size_t event_size);

#ifdef __cplusplus
}
#endif

#endif
