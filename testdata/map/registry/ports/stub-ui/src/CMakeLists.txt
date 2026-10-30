cmake_minimum_required(VERSION 3.25)
project(StubUI VERSION 1.0.0 LANGUAGES C)

add_library(stub_ui stub_ui.c)
add_library(StubUI::StubUI ALIAS stub_ui)
target_include_directories(stub_ui PUBLIC $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include> $<INSTALL_INTERFACE:include>)
# Built with wide ids; the define is deliberately not exported, so
# consumers must match it themselves (the build-flags contract's concern).
target_compile_definitions(stub_ui PRIVATE STUB_UI_WIDE_IDS=1)
set_target_properties(stub_ui PROPERTIES EXPORT_NAME StubUI)

include(GNUInstallDirs)
install(TARGETS stub_ui EXPORT StubUITargets)
install(FILES include/stub_ui.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR})
install(EXPORT StubUITargets NAMESPACE StubUI:: DESTINATION ${CMAKE_INSTALL_DATADIR}/StubUI)
file(WRITE ${CMAKE_CURRENT_BINARY_DIR}/StubUIConfig.cmake "include(\${CMAKE_CURRENT_LIST_DIR}/StubUITargets.cmake)\n")
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/StubUIConfig.cmake DESTINATION ${CMAKE_INSTALL_DATADIR}/StubUI)
