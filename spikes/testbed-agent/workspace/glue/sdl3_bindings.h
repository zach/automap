// Reference glue: SDL3 bindings to the allocation and diagnostics contracts (DRAFT, spike).
#pragma once

void sdl3_bind_allocation();   // memory phase; must be the first call into SDL
void sdl3_bind_diagnostics();  // diagnostics phase
void sdl3_register_probes();   // call after a smoke run, before shutdown
