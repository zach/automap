// Acme's engine context: what its glue shares between entry points.
#pragma once

struct stub_window;

namespace am {

// The context every entry point takes: the fields recipes provide and read.
struct Context {
    struct {
        ::stub_window* window = nullptr;
    } stub_window;
};

} // namespace am
