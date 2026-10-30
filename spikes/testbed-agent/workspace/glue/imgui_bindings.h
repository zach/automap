// Reference glue: Dear ImGui bindings to the allocation and build-flags contracts (DRAFT, spike).
#pragma once

void imgui_bind_allocation();   // memory phase; before ImGui::CreateContext()
void imgui_register_probes();   // call after a smoke run, before shutdown
