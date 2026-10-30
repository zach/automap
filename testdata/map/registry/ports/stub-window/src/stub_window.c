#include "stub_window.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct stub_window {
    char title[64];
    int width, height;
};

#define QUEUE_SIZE 16

static stub_window_alloc_fn g_alloc = malloc;
static stub_window_free_fn g_free = free;
static stub_window_log_fn g_log = NULL;
static char* g_name = NULL;
static bool g_initialized = false;
static stub_window_event* g_queue = NULL;
static int g_head = 0, g_count = 0;

static void log_message(const char* message) {
    if (g_log != NULL) g_log(message);
}

void stub_window_set_allocator(stub_window_alloc_fn alloc, stub_window_free_fn release) {
    g_alloc = alloc;
    g_free = release;
}

void stub_window_get_allocator(stub_window_alloc_fn* alloc, stub_window_free_fn* release) {
    *alloc = g_alloc;
    *release = g_free;
}

void stub_window_set_name(const char* name) {
    if (g_name != NULL) g_free(g_name);
    g_name = g_alloc(strlen(name) + 1);
    strcpy(g_name, name);
}

void stub_window_set_log(stub_window_log_fn log) { g_log = log; }

void stub_window_log_test_message(void) { log_message("test message"); }

bool stub_window_init(void) {
    g_queue = g_alloc(sizeof(stub_window_event) * QUEUE_SIZE);
    if (g_queue == NULL) return false;
    g_head = g_count = 0;
    g_initialized = true;
    log_message("initialized");
    return true;
}

stub_window* stub_window_open(const char* title, int width, int height) {
    const char* headless = getenv("STUB_WINDOW_HEADLESS");
    if (!g_initialized) return NULL;
    if (headless == NULL || strcmp(headless, "1") != 0) {
        log_message("no display: set STUB_WINDOW_HEADLESS=1");
        return NULL;
    }
    stub_window* w = g_alloc(sizeof(stub_window));
    snprintf(w->title, sizeof w->title, "%s", title);
    w->width = width;
    w->height = height;
    return w;
}

void stub_window_close(stub_window* window) { g_free(window); }

void stub_window_shutdown(void) {
    g_free(g_queue);
    g_queue = NULL;
    if (g_name != NULL) g_free(g_name);
    g_name = NULL;
    g_initialized = false;
}

void stub_window_push_event(stub_window_event event) {
    if (g_count == QUEUE_SIZE) return;
    g_queue[(g_head + g_count) % QUEUE_SIZE] = event;
    ++g_count;
}

bool stub_window_poll(stub_window_event* event) {
    if (g_count == 0) return false;
    *event = g_queue[g_head];
    g_head = (g_head + 1) % QUEUE_SIZE;
    --g_count;
    return true;
}

bool stub_window_abi_check(size_t event_size) { return event_size == sizeof(stub_window_event); }
