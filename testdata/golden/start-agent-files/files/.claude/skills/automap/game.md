# From testbed to game

`automap start` writes a testbed: a `CMakeLists.txt` and a `main.cpp` that start the composition and run a frame loop. They are yours to replace. This guide says what Automap needs from whatever replaces them. It was written by automap 0.0.0-dev.

## The executable

The manifest's `executable` names the CMake target `automap verify` builds and runs, and its sources:

```json
"executable": { "target": "mygame", "sources": ["game/app/main.cpp"] }
```

- Rename the target and move the sources as you like; edit the manifest to match and run `automap verify --json`.
- Your `CMakeLists.txt` includes `cmake/automap/automap.cmake` once and links the executable to `automap_glue`. Add your own libraries, tests and other executables beside it; `verify` builds only the manifest's target.
- The target can be defined in any directory of your CMake project.

`verify` runs the executable twice, headless, and expects it to exit 0 within two minutes:

- `--frames <n>`: run `n` frames, then shut down cleanly.
- `--probe <file>`: also run the probes and write their results to `<file>`.

A `main` that honours both, as the testbed's does:

```cpp
#include "automap/probe.h"
#include "automap/startup.h"

int main(int argc, char** argv) {
    int frames = -1;                 // -1: run until the window closes
    const char* results = nullptr;
    for (int i = 1; i + 1 < argc; ++i) {
        if (std::strcmp(argv[i], "--frames") == 0) frames = std::atoi(argv[i + 1]);
        if (std::strcmp(argv[i], "--probe") == 0) results = argv[i + 1];
    }
    if (results != nullptr) am::probe::begin(results);
    am::Context ctx;
    if (!am::startup(ctx)) {         // calls every library's setup, in order
        am::probe::write_results();
        return 1;
    }
    for (int frame = 0; frames < 0 || frame < frames; ++frame) {
        if (!game_frame(ctx)) break; // your frame
    }
    if (results != nullptr) am::probe::run(am::probe::When::running, ctx);
    am::shutdown(ctx);
    if (results != nullptr) {
        am::probe::run(am::probe::When::after_shutdown, ctx);
        am::probe::write_results();
    }
    return 0;
}
```

The run has no display and no audio device: `verify` sets what each library's facts say it needs to run headless (`automap show library <id> --json`, `facts.platform.headless`). A game that needs a file, a save directory or a network at startup must cope without them in that run.

## Where it runs, and assets

`verify` runs the executable from its own directory in the build tree (`build/automap/<platform>/`, or the subdirectory your CMake puts it in), never from the repo. Copy assets beside the executable after it builds, and load them relative to the executable:

```cmake
add_custom_command(TARGET mygame POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy_directory ${CMAKE_SOURCE_DIR}/assets $<TARGET_FILE_DIR:mygame>/assets)
```

## Game code and the glue

Game code calls each library's own API, and Automap's shared interfaces for what crosses libraries: allocation, logging and profiling, files and jobs, in `glue/automap/` (`#include "automap/<name>.h"`). Code that goes through them keeps working when the library behind a slot changes or is removed.

Everything under `glue/` and `cmake/automap/` is yours to change, except the files the CLI writes from the manifest (`choices.h`, `context.h`, `startup.h`, `startup.cpp`, `automap.cmake`). The recipe's notes say what each piece of glue is for: `automap show recipe <id> --json`. A file you change is left alone by `automap upgrade`, which puts a newer reference beside it and warns.

## Shaders

SDL_shadercross's glue takes SPIR-V and turns it into what the platform's GPU backend needs at startup. Compile shader source to SPIR-V when you build: where SDL_shadercross is in the composition, vcpkg has installed the `dxc` compiler under `build/automap/<platform>/vcpkg_installed/<triplet>/tools/directx-dxc/` (Windows and x64 Linux), and `dxc -spirv -T vs_6_0 -E main shader.hlsl -Fo shader.spv` writes SPIR-V. Commit the SPIR-V, or compile it in a build step.

## A release build

`verify` builds Debug, to probe. A release is an ordinary CMake build with vcpkg's toolchain: take the configure command from the first line of `build/automap/build-<platform>.log` and change the build directory and type.

```console
$ cmake -S . -B build/release -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_TOOLCHAIN_FILE=<vcpkg>/scripts/buildsystems/vcpkg.cmake
$ cmake --build build/release --target mygame
```

`<vcpkg>` is `VCPKG_ROOT`, or the vcpkg the CLI keeps in its cache, which that log line names.

## What to ship

The executable, your assets, and the shared libraries beside the executable in the release build directory: on Windows, vcpkg copies every DLL the executable needs there when it links. List what the executable imports before you package (`dumpbin /dependents mygame.exe` on Windows, `ldd` on Linux) and ship exactly that; don't ship `vcpkg_installed/`. Ship each library's license: `vcpkg_installed/<triplet>/share/<port>/copyright`.

A library that should not reach players, such as a profiler that records from startup, comes out of the composition before the release build: see **Changing a composition** in [`SKILL.md`](SKILL.md).
