# Acme Tools, an in-house library.
# SPDX-License-Identifier: MIT-0
add_library(acme_tools STATIC acme_tools.cpp)
add_library(Acme::Tools ALIAS acme_tools)
target_include_directories(acme_tools PUBLIC include)
target_compile_features(acme_tools PUBLIC cxx_std_17)
