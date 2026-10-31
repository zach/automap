package composition_test

import (
	"reflect"
	"testing"

	"github.com/zach/automap/internal/composition"
)

func TestScanCMake(t *testing.T) {
	p := composition.ScanCMake(map[string]string{
		"CMakeLists.txt": `
find_package(SDL3 CONFIG REQUIRED)
find_package(Threads REQUIRED) # module mode: CMake's own
add_subdirectory(engine/jobs)
# target_link_libraries(game PRIVATE Commented::Out)
add_executable(game WIN32 src/main.cpp src/game.cpp)
target_link_libraries(game PRIVATE SDL3::SDL3 Acme::Jobs Threads::Threads
    Unknown::Thing)
`,
		"engine/jobs/CMakeLists.txt": `
add_library(acme_jobs STATIC jobs.cpp)
add_library(Acme::Jobs ALIAS acme_jobs)
`,
	})

	want := []composition.ProjectExecutable{{Target: "game", Sources: []string{"src/main.cpp", "src/game.cpp"}}}
	if !reflect.DeepEqual(p.Executables, want) {
		t.Errorf("executables %+v, want %+v", p.Executables, want)
	}
	if got := p.LibraryTargets(); !reflect.DeepEqual(got, []string{"Acme::Jobs", "SDL3::SDL3"}) {
		t.Errorf("library targets %v, want the alias Acme::Jobs and the config package's SDL3::SDL3", got)
	}
}
