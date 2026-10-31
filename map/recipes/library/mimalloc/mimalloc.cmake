# mimalloc's build fragment. The port installs mimalloc-static on static
# triplets and mimalloc on dynamic ones.
# SPDX-License-Identifier: MIT-0
find_package(mimalloc CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC $<IF:$<TARGET_EXISTS:mimalloc-static>,mimalloc-static,mimalloc>)
