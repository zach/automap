# Acme Tools' build fragment. The library is in the repo: automap.cmake adds
# its directory before this fragment runs.
# SPDX-License-Identifier: MIT-0
target_link_libraries(automap_glue PUBLIC Acme::Tools)
