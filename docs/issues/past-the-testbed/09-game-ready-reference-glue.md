# 09: Game-ready reference glue

**What to build:** SDL3's setup takes the window's title and size from the context, and the Dear ImGui pairing offers its frame in pieces, so a game uses both unchanged. Field notes [33, 34].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] `ctx.sdl3.window_title`, `window_width` and `window_height`, set before `am::startup`, give the window its title and size; unset, the testbed's values apply.
- [ ] The pairing glue offers `am_pair_imgui_sdl3_process_event`, `am_pair_imgui_sdl3_new_frame` and `am_pair_imgui_sdl3_render` (into a command buffer and target the game owns); `begin_frame` and `end_frame` remain, built from them.
- [ ] The recipes' adaptation notes describe both; the pairing's probe and seeded violation still pass and are caught.
- [ ] Verify map on the branch is green for the SDL3 and pairing edges on all three platforms.
