<!-- automap:begin -->
## Automap

- Build: `cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_C_COMPILER=clang -DCMAKE_CXX_COMPILER=clang++ -DAUTOMAP_HEADLESS=ON && cmake --build build`
- Test (smoke): `VK_DRIVER_FILES=/usr/share/vulkan/icd.d/lvp_icd.json ./build/testbed --probe`
- Probe: the same run must print `"verdict":"pass"`; and, because the jobs contract is in use, a ThreadSanitizer build must run clean:
  `cmake -S . -B build-tsan -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_C_COMPILER=clang -DCMAKE_CXX_COMPILER=clang++ -DAUTOMAP_HEADLESS=ON -DCMAKE_C_FLAGS=-fsanitize=thread -DCMAKE_CXX_FLAGS=-fsanitize=thread -DCMAKE_EXE_LINKER_FLAGS=-fsanitize=thread && cmake --build build-tsan && LP_NUM_THREADS=0 VK_DRIVER_FILES=/usr/share/vulkan/icd.d/lvp_icd.json TSAN_OPTIONS=exitcode=66 ./build-tsan/testbed --probe`
  (`LP_NUM_THREADS=0` keeps Mesa's uninstrumented software rasterizer single-threaded; without it TSAN reports races inside `libvulkan_lvp.so`.)
- Manifest: `automap.manifest.json`. Lock: `automap.lock.json`. Evidence: `evidence/`.
- Slots in use: windowing=sdl3, input=sdl3, gpu=sdl3, tools-ui=dear-imgui, physics-3d=jolt.
- Go through the router skill and CLI before changing any glue.
<!-- automap:end -->
