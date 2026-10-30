# stub-physics's build fragment.
find_package(StubPhysics CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC StubPhysics::StubPhysics)
