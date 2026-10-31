# 26: miniaudio

**What to build:** miniaudio fills the audio slot, with its allocation callbacks and logging routed to the facades and a null backend for headless runs.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] A composition with miniaudio reaches probed on all three platforms without an audio device.
- [x] Every new probe fails its seeded violation.

## Comments

- miniaudio 0.11.25 is the curated port at the map's baseline. The port installs `miniaudio.h` alone, so the build fragment defines `miniaudio::miniaudio` and the map records an empty CMake package (SPEC §5.3 says so).
- Setup glue is the one translation unit with `MINIAUDIO_IMPLEMENTATION`; it owns the context and engine, which copy the allocation callbacks and the log the bindings fill in during the memory and diagnostics phases.
- miniaudio's headless platform fact, `AUTOMAP_MINIAUDIO_BACKEND=null`, makes setup glue use only the null backend, so verify needs no audio device.
- The diagnostics binding trims the newline miniaudio ends its messages with.
