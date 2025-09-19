package semver

import (
	"strings"
	"testing"
)

// helper to build Semver with flags derived from presence of parts.
func mk(valid bool, original string, hasV bool, major, minor, patch int, pre, build string) *Semver {
	var f Flags
	if hasV {
		f |= FlagHasV
	}

	if (valid) {
		f |= FlagHasMajor
	}

	// derive minor/patch presence from Original form
	if len(original) > 0 {
		dotCount := 0
		for i := 0; i < len(original); i++ {
			if original[i] == '.' {
				dotCount++
			}
		}
		if dotCount >= 1 {
			f |= FlagHasMinor
		}
		if dotCount >= 2 {
			f |= FlagHasPatch
		}
	}

	if pre != "" {
		f |= FlagHasPre
	}

	if build != "" {
		f |= FlagHasBuild
	}

	return &Semver{
		Valid:      valid,
		Original:   original,
		Flags:      f,
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: pre,
		Build:      build,
	}
}

func TestPrint_Variants(t *testing.T) {
	tests := []struct {
		name   string
		v      *Semver
		mask   PrintFlags
		expect string
	}{
		{
			name:   "Invalid",
			v:      mk(false, "invalid", false, 0, 0, 0, "", ""),
			mask:   PrintMaskDefault,
			expect: "",
		},
		{
			name:   "Blank",
			v:      mk(true, "0", false, 0, 0, 0, "", ""),
			mask:   PrintMaskDefault,
			expect: "0.0.0",
		},
		{
			name:   "Canonical: vMAJOR.MINOR.PATCH[-pre], build stripped",
			v:      mk(true, "v1.2.3-alpha+001", true, 1, 2, 3, "alpha", "001"),
			mask:   PrintMaskCanonical,
			expect: "v1.2.3-alpha",
		},
		{
			name:   "SemVer: MAJOR.MINOR.PATCH[-pre][+build], no v",
			v:      mk(true, "v1.2.3-alpha+001", true, 1, 2, 3, "alpha", "001"),
			mask:   PrintMaskSemVer,
			expect: "1.2.3-alpha+001",
		},
		{
			name:   "Default preserve uppercase V with build and pre",
			v:      mk(true, "V1.2.3-alpha+meta", true, 1, 2, 3, "alpha", "meta"),
			mask:   PrintMaskDefault,
			expect: "V1.2.3-alpha+meta",
		},
		{
			name:   "Zero-fill MINOR/PATCH when requested by release mask (only major present)",
			v:      mk(true, "1", false, 1, 0, 0, "", ""),
			mask:   PrintMaskRelease | PrintPrefixNoV, // no prefix to make expectation stable
			expect: "1.0.0",
		},
		{
			name:   "Zero-fill when only MINOR present (PATCH -> 0)",
			v:      mk(true, "1.2", false, 1, 2, 0, "", ""),
			mask:   PrintMaskRelease | PrintPrefixNoV,
			expect: "1.2.0",
		},
		{
			name:   "Prefix 'v' with only MAJOR requested",
			v:      mk(true, "1", false, 1, 0, 0, "", ""),
			mask:   PrintPrefixV | PrintMajor,
			expect: "v1",
		},
		{
			name:   "Request PATCH only -> promote MINOR and MAJOR, zero-fill",
			v:      mk(true, "1", false, 1, 0, 0, "", ""),
			mask:   PrintPrefixNoV | PrintPatch,
			expect: "1.0.0",
		},
		{
			name:   "Prerelease flag set but value empty -> no '-' emitted",
			v:      mk(true, "1.2.3", false, 1, 2, 3, "", ""),
			mask:   PrintPrefixNoV | PrintMaskRelease | PrintPrerelease,
			expect: "1.2.3",
		},
		{
			name:   "Build flag set but value empty -> no '+' emitted",
			v:      mk(true, "1.2.3", false, 1, 2, 3, "", ""),
			mask:   PrintPrefixNoV | PrintMaskRelease | PrintBuild,
			expect: "1.2.3",
		},
		{
			name:   "Explicit v prefix overrides preserve (original had no v)",
			v:      mk(true, "1.2.3-alpha+1", false, 1, 2, 3, "alpha", "1"),
			mask:   PrintPrefixV | PrintMaskRelease | PrintPrerelease,
			expect: "v1.2.3-alpha",
		},
		{
			name:   "Explicit no-v prefix overrides preserve (original had v)",
			v:      mk(true, "v1.2.3+meta", true, 1, 2, 3, "", "meta"),
			mask:   PrintPrefixNoV | PrintMaskRelease | PrintBuild,
			expect: "1.2.3+meta",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.v.Print(tc.mask)
			if got != tc.expect {
				t.Fatalf("Print() = %q, want %q", got, tc.expect)
			}
		})
	}
}

// TestMajor checks that MajorStr() returns the correct "vMAJOR".
func TestMajor(t *testing.T) {
	for _, tt := range tests {
		v, _ := Parse(tt.in)
		want := ""
		if tt.out != "" {
			if i := strings.Index(tt.out, "."); i >= 0 {
				want = tt.out[:i] // "v1"
			}
		}
		if v.Valid {
			if got := v.MajorStr(); got != want {
				t.Errorf("MajorStr(%q) = %q, want %q", tt.in, got, want)
			}
		} else if want != "" {
			t.Errorf("MajorStr(%q) invalid but want %q", tt.in, want)
		}
	}
}

// TestMajorMinor checks that MajorMinorStr() returns the correct "vMAJOR.MINOR".
func TestMajorMinor(t *testing.T) {
	for _, tt := range tests {
		v, _ := Parse(tt.in)
		var want string
		if tt.out != "" {
			want = tt.in
			if i := strings.Index(want, "+"); i >= 0 {
				want = want[:i]
			}
			if i := strings.Index(want, "-"); i >= 0 {
				want = want[:i]
			}
			switch strings.Count(want, ".") {
			case 0:
				want += ".0"
			case 1:
				// ok
			case 2:
				want = want[:strings.LastIndex(want, ".")]
			}
			// ensure leading v
			if want == "" || want[0] != 'v' {
				want = "v" + strings.TrimPrefix(want, "v")
			}
		}
		if v.Valid {
			if got := v.MajorMinorStr(); got != want {
				t.Errorf("MajorMinorStr(%q) = %q, want %q", tt.in, got, want)
			}
		} else if want != "" {
			t.Errorf("MajorMinorStr(%q) invalid but want %q", tt.in, want)
		}
	}
}

func TestBuildOriginal_SemVer(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"-", ""},
		{"someversion", ""},
		{"v1", "1.0.0"},
		{"1.2.3", "1.2.3"},
		{"1.2.3+meta", "1.2.3+meta"},
		{"v1.2.3-rc.1", "1.2.3-rc.1"},
		{"V1.2.3-rc.1+z", "1.2.3-rc.1+z"},
	}
	for _, tc := range cases {
		v, _ := Parse(tc.in)
		if got := v.SemVer(); got != tc.want {
			t.Errorf("SemVer() %q = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildOriginal_StrictV(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"1.2.3-rc.1+meta", "v1.2.3-rc.1+meta"},
		{"v1.2.3+meta", "v1.2.3+meta"},
		{"V1.2.3", "v1.2.3"},
	}
	for _, tc := range cases {
		v, _ := Parse(tc.in)
		if got := v.Full(true); got != tc.want {
			t.Errorf("Full(true) %q = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildOriginal_PreservePrefix(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"1.2.3+meta", "1.2.3+meta"},
		{"v1.2.3-rc.1", "v1.2.3-rc.1"},
		{"V1.2.3-rc.1+z", "V1.2.3-rc.1+z"},
	}
	for _, tc := range cases {
		v, _ := Parse(tc.in)
		if got := v.Full(false); got != tc.want {
			t.Errorf("Full(false) %q = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestFull covers various combinations of core/prerelease/build (and shorthand).
func TestFull(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1.2.3", "v1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"1.2.3-alpha.1+build.5", "v1.2.3-alpha.1+build.5"},
		{"v1.2.3+meta", "v1.2.3+meta"},
		{"v1.2.3-rc.1", "v1.2.3-rc.1"},
		{"1", "v1.0.0"},   // shorthand MAJOR
		{"1.2", "v1.2.0"}, // shorthand MAJOR.MINOR
		{"v1-pre", ""},    // invalid: prerelease requires x.y.z
		{"v1.2+meta", ""}, // invalid: build requires x.y.z
		{"bad", ""},       // invalid
	}

	for _, tc := range cases {
		v, _ := Parse(tc.in)
		if got := v.Full(true); got != tc.want {
			t.Errorf("Full(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestReleaseStr ensures normalized "vMAJOR.MINOR.PATCH" output.
func TestReleaseStr(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1.2.3", "v1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"1.2.3-alpha.1+build.5", "v1.2.3"},
		{"v1.2.3+meta", "v1.2.3"},
		{"v1.2.3-rc.1", "v1.2.3"},
		{"1", "v1.0.0"},   // shorthand MAJOR -> normalized
		{"1.2", "v1.2.0"}, // shorthand MAJOR.MINOR -> normalized
		{"v1-pre", ""},    // invalid overall -> empty
		{"bad", ""},       // invalid
	}

	for _, tc := range cases {
		v, _ := Parse(tc.in)
		if got := v.ReleaseStr(); got != tc.want {
			t.Errorf("ReleaseStr(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
