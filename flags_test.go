package semver

import "testing"

// TestFlags validates the presence flags and IsRelease()/HasV() across inputs.
func TestFlags(t *testing.T) {
	tests := []struct {
		in                  string
		valid               bool
		hasV                bool
		hasMajor, hasMinor  bool
		hasPatch, hasPre    bool
		hasBuild, isRelease bool
	}{
		// shorthands (accepted)
		{"1", true, false, true, false, false, false, false, true},
		{"v1", true, true, true, false, false, false, false, true},
		{"1.2", true, false, true, true, false, false, false, true},

		// full core
		{"1.2.0", true, false, true, true, true, false, false, true},
		{"1.2.3", true, false, true, true, true, false, false, true},
		{"V1.2.3", true, true, true, true, true, false, false, true},

		// prerelease / build (require full x.y.z)
		{"1.2.3-rc.1", true, false, true, true, true, true, false, false},
		{"1.2.3+build.5", true, false, true, true, true, false, true, false},
		{"1.2.3-rc.1+build.5", true, false, true, true, true, true, true, false},

		// invalid forms: pre/build without full patch
		{"v1-pre", false, true, false, false, false, false, false, false},
		{"1.2-pre", false, false, false, false, false, false, false, false},
		{"1.2+meta", false, false, false, false, false, false, false, false},

		// random invalid
		{"bad", false, false, false, false, false, false, false, false},
	}

	for _, tc := range tests {
		v, _ := Parse(tc.in)

		if v.Valid != tc.valid {
			t.Errorf("Valid(%q) = %v, want %v", tc.in, v.Valid, tc.valid)
		}

		// For invalid versions flags are expected to be zero;
		// we still assert helpers to be consistent.
		if got := v.HasV(); got != tc.hasV {
			t.Errorf("HasV(%q) = %v, want %v", tc.in, got, tc.hasV)
		}
		if got := v.HasMajor(); got != tc.hasMajor {
			t.Errorf("HasMajor(%q) = %v, want %v", tc.in, got, tc.hasMajor)
		}
		if got := v.HasMinor(); got != tc.hasMinor {
			t.Errorf("HasMinor(%q) = %v, want %v", tc.in, got, tc.hasMinor)
		}
		if got := v.HasPatch(); got != tc.hasPatch {
			t.Errorf("HasPatch(%q) = %v, want %v", tc.in, got, tc.hasPatch)
		}
		if got := v.HasPre(); got != tc.hasPre {
			t.Errorf("HasPre(%q) = %v, want %v", tc.in, got, tc.hasPre)
		}
		if got := v.HasBuild(); got != tc.hasBuild {
			t.Errorf("HasBuild(%q) = %v, want %v", tc.in, got, tc.hasBuild)
		}
		if got := v.IsRelease(); got != tc.isRelease {
			t.Errorf("IsRelease(%q) = %v, want %v", tc.in, got, tc.isRelease)
		}
	}
}
