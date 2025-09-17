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
		if got := v.Canonical(); got != tt.out {
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
		if got := v.Canonical(); got != tt.out {
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
				want = tt.out[i+1:] // without '-'
			}
		}
		if got := v.Prerelease; got != want {
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
				want = tt.in[i+1:] // without '+'
			}
		}
		if got := v.Build; got != want {
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

var benchInputs = []string{
	"1.2.3",
	"v10.20.30",
	"2.0.0-rc.1",
	"3.4.5-alpha.7",
	"6.7.8+build.11",
	"9.9.9-beta",
	"0.0.1",
	"1.0.0+meta-pre.sha.256a",
	"4.5.6-zzz",
	"7.8.9-1",
}

// that the compiler does not throw out calls
var sinkInt int
var sinkStr string

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

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	n := 0
	for i := 0; i < b.N; i++ {
		for _, s := range benchInputs {
			v, ok := Parse(s)
			if ok {
				n += v.Major
			}
		}
	}
	sinkInt = n
}

// Benchmark full Compare() on versions that *hit* comparePrerelease path.
func BenchmarkCompare_PreRelease(b *testing.B) {
	cases := []struct {
		name string
		a, b string
	}{
		{"Equal", "1.2.3-alpha.1", "1.2.3-alpha.1"},              // equal prerelease
		{"NumOrderShortVsLong", "1.2.3-beta.2", "1.2.3-beta.11"}, // numeric vs numeric
		{"NumericVsAlpha", "1.2.3-alpha.1", "1.2.3-alpha.beta"},  // numeric < non-numeric
		{"Lexical", "1.2.3-alpha.beta", "1.2.3-alpha.gamma"},     // lexical ASCII
		{"DeepChain", "1.2.3-a.10.b.2", "1.2.3-a.2.b.10"},        // deeper chain
		{"FirstIdentDiff", "1.2.3-alpha", "1.2.3-beta"},          // different first identifier
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			v1, _ := Parse(tc.a)
			v2, _ := Parse(tc.b)
			b.ReportAllocs()
			b.ResetTimer()
			sum := 0
			for i := 0; i < b.N; i++ {
				sum += v1.Compare(v2)
			}
			sinkInt = sum
		})
	}
}

// Benchmark the internal comparePrerelease() directly (isolated).
// Note: current implementation allocates due to "-" + a trick.
func BenchmarkCompare_PreRelease_Direct(b *testing.B) {
	cases := []struct {
		name string
		a, b string
	}{
		{"Equal", "alpha.1", "alpha.1"},
		{"NumOrderShortVsLong", "beta.2", "beta.11"},
		{"NumericVsAlpha", "alpha.1", "alpha.beta"},
		{"Lexical", "alpha.beta", "alpha.gamma"},
		{"DeepChain", "a.10.b.2", "a.2.b.10"},
		{"FirstIdentDiff", "alpha", "beta"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			sum := 0
			for i := 0; i < b.N; i++ {
				sum += comparePrerelease(tc.a, tc.b)
			}
			sinkInt = sum
		})
	}
}

// Build canonical on a plain release (build metadata stripped).
func BenchmarkCanonical_Release(b *testing.B) {
	v, _ := Parse("1.2.3+meta.whatever")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = v.Canonical()
	}
}

// Build canonical when prerelease is present.
func BenchmarkCanonical_Prerelease(b *testing.B) {
	v, _ := Parse("1.2.3-alpha.1")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = v.Canonical()
	}
}

// Longer prerelease chain stresses byte appends.
func BenchmarkCanonical_LongPrerelease(b *testing.B) {
	v, _ := Parse("10.20.30-alpha.beta.1.2.3.4.5.6.7.8.9-rc.1+build.123") // build is stripped
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = v.Canonical()
	}
}

// String() should be identical to Canonical() by implementation.
func BenchmarkString_AliasOfCanonical(b *testing.B) {
	v, _ := Parse("3.4.5-rc.2+build.9")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = v.String()
	}
}

// End-to-end cost: Parse + Canonical in each iteration.
func BenchmarkParseThenCanonical(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	n := 0
	for i := 0; i < b.N; i++ {
		for _, s := range benchInputs {
			v, ok := Parse(s)
			if ok {
				sinkStr = v.Canonical()
				// consume a bit to avoid over-optimizations
				n += len(sinkStr)
			}
		}
	}
	if n == 42 { // make sure n is used
		b.Log(n)
	}
}
