/* stub-physics: a tiny physics library for Automap's fixture map. Each step
 * splits its work into tasks; with job hooks installed it submits them to
 * the engine's scheduler, otherwise it runs them inline. MIT-0. */
#ifndef STUB_PHYSICS_H
#define STUB_PHYSICS_H

#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef void (*stub_physics_task_fn)(void* data);
/* Submits a task; the library later waits for every task it submitted. */
typedef void (*stub_physics_submit_fn)(stub_physics_task_fn task, void* data, void* user);
typedef void (*stub_physics_wait_fn)(void* user);

void stub_physics_set_job_hooks(stub_physics_submit_fn submit, stub_physics_wait_fn wait, void* user);

bool stub_physics_create_world(int bodies);
/* Advances every body by one step, in parallel through the job hooks. */
void stub_physics_step(void);
/* The sum of every body's position, for checking a step ran. */
long long stub_physics_checksum(void);
void stub_physics_destroy_world(void);

#ifdef __cplusplus
}
#endif

#endif
