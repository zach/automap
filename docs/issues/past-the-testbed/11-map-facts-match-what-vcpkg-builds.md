# 11: Map facts match what vcpkg builds

**What to build:** SDL3's default variant states the features vcpkg builds, and Automap's CI compares every reference composition's install plan with its variants' stated features. Field notes [14].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] SDL3's default variant names `ibus` among its Linux features, its description says IBus is on and why, and the system dependencies still hold.
- [ ] A check (`ci/cmd/verifymap plan-check` or a test in the public-map CI job) runs `vcpkg install --dry-run` for each reference composition and fails when a library's port has a feature in the plan that the variant doesn't state or imply through the port's own feature dependencies.
- [ ] The check passes on `main` and fails when SDL3's variant drops `ibus`.
- [ ] CI runs it on every push, on Linux.
