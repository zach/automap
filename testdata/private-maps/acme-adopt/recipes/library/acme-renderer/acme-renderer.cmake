# Acme Renderer's build fragment. The library is in the repo: automap.cmake adds
# its directory, unless the project already has, before this fragment runs.
# SPDX-License-Identifier: MIT-0
target_link_libraries(automap_glue PUBLIC Acme::Renderer)
