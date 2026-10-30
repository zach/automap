cmake_minimum_required(VERSION 3.25)
project(StubPhysics VERSION 1.0.0 LANGUAGES C)

add_library(stub_physics stub_physics.c)
add_library(StubPhysics::StubPhysics ALIAS stub_physics)
target_include_directories(stub_physics PUBLIC $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include> $<INSTALL_INTERFACE:include>)
set_target_properties(stub_physics PROPERTIES EXPORT_NAME StubPhysics)

include(GNUInstallDirs)
install(TARGETS stub_physics EXPORT StubPhysicsTargets)
install(FILES include/stub_physics.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR})
install(EXPORT StubPhysicsTargets NAMESPACE StubPhysics:: DESTINATION ${CMAKE_INSTALL_DATADIR}/StubPhysics)
file(WRITE ${CMAKE_CURRENT_BINARY_DIR}/StubPhysicsConfig.cmake "include(\${CMAKE_CURRENT_LIST_DIR}/StubPhysicsTargets.cmake)\n")
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/StubPhysicsConfig.cmake DESTINATION ${CMAKE_INSTALL_DATADIR}/StubPhysics)
