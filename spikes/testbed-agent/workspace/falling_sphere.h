// Game code: a sphere falling onto a static floor, simulated by Jolt.
#pragma once

namespace JPH { class JobSystem; }

// systems phase, after jolt_register_types(): creates the TempAllocator and PhysicsSystem.
bool falling_sphere_init_physics(JPH::JobSystem* job_system);
// game phase: creates the floor and the sphere.
void falling_sphere_create_bodies();
// Advances the simulation by dt seconds (one Jolt step), using the job system given at init.
void falling_sphere_step(float dt);
// Drops the sphere again from its start height.
void falling_sphere_reset();
// Height of the sphere's centre above the floor surface, in metres.
float falling_sphere_height();
// game phase, shutdown: removes and destroys the bodies.
void falling_sphere_destroy_bodies();
// systems phase, shutdown, before jolt_unregister_types(): destroys the PhysicsSystem and TempAllocator.
void falling_sphere_shutdown_physics();
