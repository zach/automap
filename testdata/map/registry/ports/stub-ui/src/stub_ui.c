#include "stub_ui.h"

#include <stdlib.h>
#include <string.h>

typedef struct context {
    void* window;
    bool input;
    long long pending_id, last_id;
} context;

static void* default_alloc(size_t size, void* user) {
    (void)user;
    return malloc(size);
}
static void default_release(void* block, void* user) {
    (void)user;
    free(block);
}

static stub_ui_allocator g_allocator = {default_alloc, default_release, NULL};
static stub_ui_assert_fn g_assert = NULL;
static char* g_name = NULL;
static context* g_ctx = NULL;

static void fail_assert(const char* expression, int line) {
    if (g_assert != NULL) g_assert(expression, __FILE__, line);
}

void stub_ui_set_allocator(const stub_ui_allocator* allocator) { g_allocator = *allocator; }

void stub_ui_set_name(const char* name) {
    if (g_name != NULL) g_allocator.release(g_name, g_allocator.user);
    g_name = g_allocator.alloc(strlen(name) + 1, g_allocator.user);
    strcpy(g_name, name);
}

void stub_ui_set_assert_handler(stub_ui_assert_fn handler) { g_assert = handler; }

void stub_ui_fail_test_assertion(void) { fail_assert("test assertion", __LINE__); }

bool stub_ui_create_context(void) {
    g_ctx = g_allocator.alloc(sizeof(context), g_allocator.user);
    if (g_ctx == NULL) return false;
    memset(g_ctx, 0, sizeof *g_ctx);
    g_ctx->last_id = -1;
    return true;
}

void stub_ui_destroy_context(void) {
    if (g_ctx != NULL && g_ctx->window != NULL) fail_assert("detached before destroy", __LINE__);
    g_allocator.release(g_ctx, g_allocator.user);
    g_ctx = NULL;
    if (g_name != NULL) g_allocator.release(g_name, g_allocator.user);
    g_name = NULL;
}

bool stub_ui_attach(void* native_window) {
    if (g_ctx == NULL || native_window == NULL) return false;
    g_ctx->window = native_window;
    return true;
}

void stub_ui_detach(void) {
    if (g_ctx != NULL) g_ctx->window = NULL;
}

void stub_ui_enable_input(void) {
    if (g_ctx == NULL || g_ctx->window == NULL) {
        fail_assert("input enabled before attach", __LINE__);
        return;
    }
    g_ctx->input = true;
}

void stub_ui_disable_input(void) {
    if (g_ctx != NULL) g_ctx->input = false;
}

void stub_ui_feed_event(long long id, int x, int y) {
    (void)x;
    (void)y;
    if (g_ctx != NULL && g_ctx->input) g_ctx->pending_id = id;
}

void stub_ui_new_frame(void) {
    if (g_ctx == NULL) return;
    g_ctx->last_id = g_ctx->pending_id;
}

long long stub_ui_last_event_id(void) { return g_ctx == NULL ? -1 : g_ctx->last_id; }

bool stub_ui_abi_check(int id_bits) { return id_bits == STUB_UI_ID_BITS; }
