# nlohmann/json's build fragment: header-only, from the nlohmann-json port.
# SPDX-License-Identifier: MIT-0
find_package(nlohmann_json CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC nlohmann_json::nlohmann_json)
