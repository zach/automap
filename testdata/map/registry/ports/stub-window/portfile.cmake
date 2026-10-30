# The source ships inside the port, so the fixture registry needs no network.
vcpkg_check_features(OUT_FEATURE_OPTIONS FEATURE_OPTIONS FEATURES wide-events STUB_WINDOW_WIDE_EVENTS)
vcpkg_configure_cmake(SOURCE_PATH "${CMAKE_CURRENT_LIST_DIR}/src" PREFER_NINJA OPTIONS ${FEATURE_OPTIONS})
vcpkg_install_cmake()
vcpkg_fixup_cmake_targets(CONFIG_PATH share/StubWindow TARGET_PATH share/StubWindow)
file(REMOVE_RECURSE "${CURRENT_PACKAGES_DIR}/debug/include" "${CURRENT_PACKAGES_DIR}/debug/share")
file(INSTALL "${CMAKE_CURRENT_LIST_DIR}/src/LICENSE" DESTINATION "${CURRENT_PACKAGES_DIR}/share/${PORT}" RENAME copyright)
