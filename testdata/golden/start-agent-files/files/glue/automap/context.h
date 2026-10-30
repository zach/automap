// Written by automap from the composition's recipes. Edit the manifest, not this file.
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
