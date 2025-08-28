package semver

import (
	"math/rand"
	"slices"
	"strings"
	"testing"
)

// tests is based on the original table from x/mod/semver.
// "out" is the expected Canon() result, or empty string if the version is invalid.
var tests = []struct {
	in  string
	out string
}{
	{"bad", ""},
	{"v1-alpha.beta.gamma", ""},
	{"v1-pre", ""},
	{"v1+meta", ""},
	{"v1-pre+meta", ""},
	{"v1.2-pre", ""},
	{"v1.2+meta", ""},
	{"v1.2-pre+meta", ""},
	{"v1.0.0-alpha", "v1.0.0-alpha"},
	{"v1.0.0-alpha.1", "v1.0.0-alpha.1"},
	{"v1.0.0-alpha.beta", "v1.0.0-alpha.beta"},
	{"v1.0.0-beta", "v1.0.0-beta"},
	{"v1.0.0-beta.2", "v1.0.0-beta.2"},
	{"v1.0.0-beta.11", "v1.0.0-beta.11"},
	{"v1.0.0-rc.1", "v1.0.0-rc.1"},
	{"v1", "v1.0.0"},
	{"v1.0", "v1.0.0"},
	{"v1.0.0", "v1.0.0"},
	{"v1.2", "v1.2.0"},
	{"v1.2.0", "v1.2.0"},
	{"v1.2.3-456", "v1.2.3-456"},
	{"v1.2.3-456.789", "v1.2.3-456.789"},
	{"v1.2.3-456-789", "v1.2.3-456-789"},
	{"v1.2.3-456a", "v1.2.3-456a"},
	{"v1.2.3-pre", "v1.2.3-pre"},
	{"v1.2.3-pre+meta", "v1.2.3-pre"},
	{"v1.2.3-pre.1", "v1.2.3-pre.1"},
	{"v1.2.3-zzz", "v1.2.3-zzz"},
	{"v1.2.3", "v1.2.3"},
	{"v1.2.3+meta", "v1.2.3"},
	{"v1.2.3+meta-pre", "v1.2.3"},
	{"v1.2.3+meta-pre.sha.256a", "v1.2.3"},
}

// novTests contains additional cases: input without leading "v"
// should be parsed and canonicalized as if prefixed with "v".
var novTests = []struct {
	in  string
	out string
}{
	{"1", "v1.0.0"},
	{"1.0", "v1.0.0"},
	{"1.2", "v1.2.0"},
	{"1.2.3", "v1.2.3"},
	{"1.2.3-rc.1", "v1.2.3-rc.1"},
	{"1.2.3+meta", "v1.2.3"}, // build strip
}

// TestParseAndCanon ensures Parse() and Canon() match the expected table.
func TestParseAndCanon(t *testing.T) {
	for _, tt := range tests {
		v, ok := Parse(tt.in)
		if ok != (tt.out != "") {
			t.Fatalf("Parse(%q) ok=%v, want ok=%v", tt.in, ok, !ok)
		}
		if tt.out == "" {
			continue
		}
		if !v.Valid {
			t.Fatalf("Parse(%q) -> Valid=false, want true", tt.in)
		}
		if got := v.Canon(); got != tt.out {
			t.Errorf("Canon(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

// TestParseNoVAndCanon ensures versions without a leading "v"
// are still parsed and canonicalized correctly.
func TestParseNoVAndCanon(t *testing.T) {
	for _, tt := range novTests {
		v, ok := Parse(tt.in)
		if !ok || !v.Valid {
			t.Fatalf("Parse(%q) -> invalid, want valid", tt.in)
		}
		if got := v.Canon(); got != tt.out {
			t.Errorf("Canon(%q) = %q, want %q", tt.in, got, tt.out)
		}
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

// TestPrerelease checks that Pre() returns the prerelease string without leading '-'.
func TestPrerelease(t *testing.T) {
	for _, tt := range tests {
		v, _ := Parse(tt.in)
		var want string
		if tt.out != "" {
			if i := strings.Index(tt.out, "-"); i >= 0 {
				want = tt.out[i+1:] // без '-'
			}
		}
		if got := v.Pre(); got != want {
			t.Errorf("Pre(%q) = %q, want %q", tt.in, got, want)
		}
	}
}

// TestBuild checks that BuildMeta() returns the build metadata without leading '+'.
func TestBuild(t *testing.T) {
	for _, tt := range tests {
		v, _ := Parse(tt.in)
		var want string
		if tt.out != "" {
			if i := strings.Index(tt.in, "+"); i >= 0 {
				want = tt.in[i+1:] // без '+'
			}
		}
		if got := v.BuildMeta(); got != want {
			t.Errorf("BuildMeta(%q) = %q, want %q", tt.in, got, want)
		}
	}
}

// TestCompare checks that Compare() preserves the ordering
// implied by the test table.
func TestCompare(t *testing.T) {
	for i, ti := range tests {
		vi, _ := Parse(ti.in)
		for j, tj := range tests {
			vj, _ := Parse(tj.in)

			cmp := vi.Compare(vj)

			var want int
			if ti.out == tj.out {
				want = 0
			} else if i < j {
				want = -1
			} else {
				want = +1
			}
			if cmp != want {
				t.Errorf("Compare(%q, %q) = %d, want %d", ti.in, tj.in, cmp, want)
			}
		}
	}
}

// TestSort checks that List.Sort() produces the expected golden order.
func TestSort(t *testing.T) {
	versions := make([]Semver, len(tests))
	for i, test := range tests {
		v, _ := Parse(test.in)
		v.Original = test.in
		versions[i] = v
	}

	rand.Shuffle(len(versions), func(i, j int) { versions[i], versions[j] = versions[j], versions[i] })
	List(versions).Sort()

	got := make([]string, len(versions))
	for i, v := range versions {
		got[i] = v.Original
	}

	golden := []string{
		"bad",
		"v1+meta",
		"v1-alpha.beta.gamma",
		"v1-pre",
		"v1-pre+meta",
		"v1.2+meta",
		"v1.2-pre",
		"v1.2-pre+meta",
		"v1.0.0-alpha",
		"v1.0.0-alpha.1",
		"v1.0.0-alpha.beta",
		"v1.0.0-beta",
		"v1.0.0-beta.2",
		"v1.0.0-beta.11",
		"v1.0.0-rc.1",
		"v1",
		"v1.0",
		"v1.0.0",
		"v1.2",
		"v1.2.0",
		"v1.2.3-456",
		"v1.2.3-456.789",
		"v1.2.3-456-789",
		"v1.2.3-456a",
		"v1.2.3-pre",
		"v1.2.3-pre+meta",
		"v1.2.3-pre.1",
		"v1.2.3-zzz",
		"v1.2.3",
		"v1.2.3+meta",
		"v1.2.3+meta-pre",
		"v1.2.3+meta-pre.sha.256a",
	}
	if !slices.Equal(got, golden) {
		t.Errorf("list is not sorted correctly\ngot:\n%v\nwant:\n%v", got, golden)
	}
}

// BenchmarkCompare benchmarks Compare() between two versions
// that differ only in build metadata (should compare equal).
func BenchmarkCompare(b *testing.B) {
	v1, _ := Parse("v1.0.0+metadata-dash")
	v2, _ := Parse("v1.0.0+metadata-dash1")
	for i := 0; i < b.N; i++ {
		if v1.Compare(v2) != 0 {
			b.Fatalf("bad compare")
		}
	}
}
