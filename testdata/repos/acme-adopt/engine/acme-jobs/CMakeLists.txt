# Acme Jobs, an in-house library.
# SPDX-License-Identifier: MIT-0
add_library(acme_jobs STATIC acme_jobs.cpp)
add_library(Acme::Jobs ALIAS acme_jobs)
target_include_directories(acme_jobs PUBLIC include)
target_compile_features(acme_jobs PUBLIC cxx_std_17)
