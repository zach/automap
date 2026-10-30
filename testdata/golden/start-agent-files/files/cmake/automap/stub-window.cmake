# stub-window's build fragment: finds the package and links it into the glue.
find_package(StubWindow CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC StubWindow::StubWindow)
