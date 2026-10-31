# enkiTS's build fragment.
# SPDX-License-Identifier: MIT-0
find_package(enkiTS CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC enkiTS::enkiTS)
