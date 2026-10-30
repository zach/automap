#include "stub_physics.h"

#include <stdlib.h>

#define TASKS 8

typedef struct task {
    long long* positions;
    int first, count;
} task;

static stub_physics_submit_fn g_submit = NULL;
static stub_physics_wait_fn g_wait = NULL;
static void* g_user = NULL;
static long long* g_positions = NULL;
static int g_bodies = 0;
static task g_tasks[TASKS];

static void advance(void* raw) {
    task* t = raw;
    for (int i = t->first; i < t->first + t->count; ++i) t->positions[i] += 1;
}

void stub_physics_set_job_hooks(stub_physics_submit_fn submit, stub_physics_wait_fn wait, void* user) {
    g_submit = submit;
    g_wait = wait;
    g_user = user;
}

bool stub_physics_create_world(int bodies) {
    g_positions = calloc((size_t)bodies, sizeof(long long));
    g_bodies = bodies;
    return g_positions != NULL;
}

void stub_physics_step(void) {
    int per = (g_bodies + TASKS - 1) / TASKS;
    for (int i = 0; i < TASKS; ++i) {
        int first = i * per;
        int count = first >= g_bodies ? 0 : (first + per > g_bodies ? g_bodies - first : per);
        g_tasks[i] = (task){g_positions, first, count};
        if (g_submit != NULL) g_submit(advance, &g_tasks[i], g_user);
        else advance(&g_tasks[i]);
    }
    if (g_wait != NULL) g_wait(g_user);
}

long long stub_physics_checksum(void) {
    long long sum = 0;
    for (int i = 0; i < g_bodies; ++i) sum += g_positions[i];
    return sum;
}

void stub_physics_destroy_world(void) {
    free(g_positions);
    g_positions = NULL;
    g_bodies = 0;
}
