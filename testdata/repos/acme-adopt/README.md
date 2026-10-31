# Acme's engine, to adopt

A studio's hand-built engine, after the product spec's Acme example, with stub
sources so it builds: the fixture map's stub-window and stub-physics from
vcpkg, an in-house scheduler (`engine/acme-jobs`, `Acme::Jobs`) and renderer
(`engine/acme-renderer`, `Acme::Renderer`), glue of its own in `engine/glue/`,
and its own startup in `engine/core/engine_main.cpp`. Its glue builds through
the glue target `cmake/automap/automap.cmake` defines. No map describes the
in-house libraries until a test adds `testdata/private-maps/acme-adopt/` as
`automap-private/`. Everything here is MIT-0.
