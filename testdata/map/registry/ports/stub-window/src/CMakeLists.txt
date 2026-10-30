cmake_minimum_required(VERSION 3.25)
project(StubWindow VERSION 1.0.0 LANGUAGES C)

option(STUB_WINDOW_WIDE_EVENTS "64-bit event ids" OFF)

add_library(stub_window stub_window.c)
add_library(StubWindow::StubWindow ALIAS stub_window)
target_include_directories(stub_window PUBLIC $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include> $<INSTALL_INTERFACE:include>)
set_target_properties(stub_window PROPERTIES EXPORT_NAME StubWindow)
if(STUB_WINDOW_WIDE_EVENTS)
    target_compile_definitions(stub_window PRIVATE STUB_WINDOW_WIDE_EVENTS=1)
endif()

include(GNUInstallDirs)
install(TARGETS stub_window EXPORT StubWindowTargets)
install(FILES include/stub_window.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR})
install(EXPORT StubWindowTargets NAMESPACE StubWindow:: DESTINATION ${CMAKE_INSTALL_DATADIR}/StubWindow)
file(WRITE ${CMAKE_CURRENT_BINARY_DIR}/StubWindowConfig.cmake "include(\${CMAKE_CURRENT_LIST_DIR}/StubWindowTargets.cmake)\n")
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/StubWindowConfig.cmake DESTINATION ${CMAKE_INSTALL_DATADIR}/StubWindow)
