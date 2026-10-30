# 26: miniaudio

**What to build:** miniaudio fills the audio slot, with its allocation callbacks and logging routed to the facades and a null backend for headless runs.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] A composition with miniaudio reaches probed on all three platforms without an audio device.
- [ ] Every new probe fails its seeded violation.
