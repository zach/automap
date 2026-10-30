# stub-window's build fragment: finds the package and builds its glue.
find_package(StubWindow CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC StubWindow::StubWindow)
set_target_properties(automap_glue PROPERTIES CXX_STANDARD ${AUTOMAP_CXX_STANDARD} CXX_STANDARD_REQUIRED ON)
