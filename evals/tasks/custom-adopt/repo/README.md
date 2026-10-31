# Acme's engine

Acme Games' hand-built engine: SDL3 and Jolt from vcpkg, an in-house
scheduler (`engine/acme-jobs`, `Acme::Jobs`) and renderer
(`engine/acme-renderer`, `Acme::Renderer`), glue of its own in
`engine/glue/`, and its own startup in `engine/core/engine_main.cpp`.
