package verify_test

import (
	"reflect"
	"testing"

	"github.com/zach/automap/internal/verify"
)

func TestParsePlan(t *testing.T) {
	out := `Detecting compiler hash for triplet x64-linux...
The following packages will be built and installed:
    joltphysics[core,double-precision]:x64-linux@5.6.0 -- git+https://github.com/zach/automap@25cb15bd42d940d1ab9402e9fe03c01b44cc8b89
    sdl3[core,vulkan,wayland,x11]:x64-linux@3.4.18
    tracy[core]:x64-linux@0.14.1#1 -- /repo/automap-private/ports/tracy
  * vcpkg-cmake:x64-linux@2025-08-07 -- git+https://github.com/microsoft/vcpkg@16b9390dfeaca6df7288ba3f827e40e48e5fc942
  * vcpkg-cmake:x64-windows@2025-08-07 -- git+https://github.com/microsoft/vcpkg@16b9390dfeaca6df7288ba3f827e40e48e5fc942
Additional packages (*) will be modified to complete this operation.
`
	want := []verify.PlannedPort{
		{Name: "joltphysics", Version: "5.6.0", Repository: "https://github.com/zach/automap", Tree: "25cb15bd42d940d1ab9402e9fe03c01b44cc8b89"},
		{Name: "sdl3", Version: "3.4.18"},
		{Name: "tracy", Version: "0.14.1#1", Dir: "/repo/automap-private/ports/tracy"},
		{Name: "vcpkg-cmake", Version: "2025-08-07", Repository: "https://github.com/microsoft/vcpkg", Tree: "16b9390dfeaca6df7288ba3f827e40e48e5fc942"},
	}
	if got := verify.ParsePlan(out); !reflect.DeepEqual(got, want) {
		t.Errorf("ParsePlan:\n got %+v\nwant %+v", got, want)
	}
}
