# 09: Game-ready reference glue

**What to build:** SDL3's setup takes the window's title and size from the context, and the Dear ImGui pairing offers its frame in pieces, so a game uses both unchanged. Field notes [33, 34].

**Blocked by:** None

**Status:** ready-for-agent

- [x] `ctx.sdl3.window_title`, `window_width`, `window_height` and `window_flags`, set before `am::startup`, give the window its title, size and flags; unset, the testbed's values apply.
- [x] The pairing glue offers `am_pair_imgui_sdl3_process_event`, `am_pair_imgui_sdl3_new_frame` and `am_pair_imgui_sdl3_render` (into a command buffer and target the game owns), and `prepare` and `draw` for ImGui inside one of the game's own passes; `begin_frame` and `end_frame` remain, built from them.
- [x] The recipes' adaptation notes and `game.md` describe both; the pairing's probe checks the pieces a game calls, and its seeded violations apply to the new glue.
- [ ] Verify map on the branch is green for the SDL3 and pairing edges on all three platforms.
