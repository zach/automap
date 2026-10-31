# Reference repo

A developer's repo as `automap start --with sdl3 --with dear-imgui:docking` wrote it, against map release `map-2026.10.31`. Automap's CI runs the reference CI workflow (`.github/workflows/automap-verify.yml`) on it, as a developer's repo would call it, so the workflow is known to produce signed composition evidence on Windows, Linux and macOS; the evidence lands in `evidence/` beside the lock.
