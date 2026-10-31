# 33: Per-platform variants, NDA maps and NDA signing

**What to build:** A manifest chooses a library's variant once or per platform, and the lock pins each platform whose variant differs. An NDA platform's private map adds console variants and its platform to libraries' targets. Evidence for an NDA platform is signed with the holder's own key and never written to a public transparency log (ADR-0048, ADR-0049, ADR-0075).

**Blocked by:** [17: Developer CI signing and the reference workflow](17-developer-ci-signing-and-reference-workflow.md), [31: Private maps, in-repo libraries and private variants](31-private-maps-and-in-repo-libraries.md)

**Status:** wontfix

- [ ] The Nebula examples from the product spec resolve and validate, with per-platform pins exactly where variants differ.
- [ ] `verify --platform <nda platform>` signs with a configured key, writes no transparency-log entry, and stores evidence in that platform's private map.
- [ ] The public map and Automap's CI never contain NDA platform names.

## Comments

NDA handling is deferred indefinitely ([ADR-0083](../../adr/0083-nda-platforms-are-deferred.md), OQ-012). Per-platform variant choices and pins already resolve and lock; the NDA maps, NDA signing and the Nebula criteria are not built.
