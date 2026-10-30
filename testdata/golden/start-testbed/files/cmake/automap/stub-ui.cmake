# stub-ui's build fragment: finds the package, builds its glue, and matches
# the ABI define stub-ui was built with.
find_package(StubUI CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC StubUI::StubUI)
target_compile_definitions(automap_glue PUBLIC STUB_UI_WIDE_IDS=1)
