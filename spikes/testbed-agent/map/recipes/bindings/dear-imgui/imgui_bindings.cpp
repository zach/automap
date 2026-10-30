// Reference glue: Dear ImGui bindings to the allocation and build-flags contracts (DRAFT, spike).
#include "imgui_bindings.h"

#include <imgui.h>

#include "automap/alloc.h"
#include "automap/lifecycle.h"
#include "automap/probe.h"

using am::alloc::Client;
using am::lifecycle::Phase;

namespace {
bool g_context_existed_at_install = true;

void* am_imgui_alloc(size_t size, void*) { return am::alloc::allocate(Client::dear_imgui, size); }
void am_imgui_free(void* block, void*) { am::alloc::release(Client::dear_imgui, block); }
} // namespace

void imgui_bind_allocation() {
    g_context_existed_at_install = ImGui::GetCurrentContext() != nullptr;
    ImGui::SetAllocatorFunctions(am_imgui_alloc, am_imgui_free, nullptr);
    am::lifecycle::init_step(Phase::memory, "dear-imgui", "ImGui::SetAllocatorFunctions");
}

void imgui_register_probes() {
    am::probe::check("allocation.hooks-before-first-allocation", "dear-imgui", !g_context_existed_at_install, "ALLOC_HOOK_LATE(dear-imgui)");

    ImGuiMemAllocFunc a = nullptr; ImGuiMemFreeFunc f = nullptr; void* ud = nullptr;
    ImGui::GetAllocatorFunctions(&a, &f, &ud);
    bool routed = a == am_imgui_alloc && f == am_imgui_free && am::alloc::routed_count(Client::dear_imgui) > 0;
    am::probe::check("allocation.routed-through-facade", "dear-imgui", routed, "ALLOC_NOT_ROUTED(dear-imgui)");

    bool layout_ok = ImGui::DebugCheckVersionAndDataLayout(IMGUI_VERSION, sizeof(ImGuiIO), sizeof(ImGuiStyle), sizeof(ImVec2),
                                                           sizeof(ImVec4), sizeof(ImDrawVert), sizeof(ImDrawIdx));
    am::probe::check("build-flags.abi-match", "dear-imgui", layout_ok, "DEFINE_MISMATCH(dear-imgui)");
}
