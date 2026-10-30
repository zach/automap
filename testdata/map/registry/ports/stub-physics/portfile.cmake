# The source ships inside the port, so the fixture registry needs no network.
# The stub libraries export no DLL symbols: build them static under every triplet.
vcpkg_check_linkage(ONLY_STATIC_LIBRARY)
vcpkg_configure_cmake(SOURCE_PATH "${CMAKE_CURRENT_LIST_DIR}/src" PREFER_NINJA)
vcpkg_install_cmake()
vcpkg_fixup_cmake_targets(CONFIG_PATH share/StubPhysics TARGET_PATH share/StubPhysics)
file(REMOVE_RECURSE "${CURRENT_PACKAGES_DIR}/debug/include" "${CURRENT_PACKAGES_DIR}/debug/share")
file(INSTALL "${CMAKE_CURRENT_LIST_DIR}/src/LICENSE" DESTINATION "${CURRENT_PACKAGES_DIR}/share/${PORT}" RENAME copyright)
