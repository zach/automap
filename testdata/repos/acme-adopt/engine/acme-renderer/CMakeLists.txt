# Acme Renderer, an in-house library.
# SPDX-License-Identifier: MIT-0
add_library(acme_renderer STATIC acme_renderer.cpp)
add_library(Acme::Renderer ALIAS acme_renderer)
target_include_directories(acme_renderer PUBLIC include)
target_compile_features(acme_renderer PUBLIC cxx_std_17)
