# The Automap registry

A vcpkg git registry (ADR-0034, ADR-0065): ports for variants vcpkg's
additive features can't express, libraries vcpkg lacks, and fresher commits.
The ports are kept on main in `ci/registry/ports/` and published here by
`ci/cmd/registry`; the public map's `map.json` names the commit its
compositions use. Everything here is MIT-0 unless a port's own files say
otherwise.
