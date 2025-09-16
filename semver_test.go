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

// Checks that ParseNoCanon matches Parse on fields,
// but does not build Canonical (Canon() == "" and String() == "").
func TestParseNoCanon_EqualsParseFields(t *testing.T) {
	for _, tt := range tests {
		v1, ok1 := Parse(tt.in)
		v2, ok2 := ParseNoCanon(tt.in)

		if ok1 != ok2 {
			t.Fatalf("ok mismatch for %q: Parse=%v ParseNoCanon=%v", tt.in, ok1, ok2)
		}
		if !ok1 {
			// обе невалидны — ок
			continue
		}

		// Числовые поля и метаданные совпадают
		if v1.Major != v2.Major || v1.Minor != v2.Minor || v1.Patch != v2.Patch {
			t.Fatalf("numeric mismatch for %q: Parse(%d.%d.%d) vs NoCanon(%d.%d.%d)",
				tt.in, v1.Major, v1.Minor, v1.Patch, v2.Major, v2.Minor, v2.Patch)
		}
		if v1.Prerelease != v2.Prerelease {
			t.Fatalf("prerelease mismatch for %q: %q vs %q", tt.in, v1.Prerelease, v2.Prerelease)
		}
		if v1.Build != v2.Build {
			t.Fatalf("build mismatch for %q: %q vs %q", tt.in, v1.Build, v2.Build)
		}
		if v2.Canon() != "" || v2.String() != "" {
			t.Fatalf("ParseNoCanon must not build Canonical for %q: got Canon=%q String=%q",
				tt.in, v2.Canon(), v2.String())
		}
		// Original должен сохраняться как есть
		if v2.Original != tt.in {
			t.Fatalf("Original mismatch for %q: got %q", tt.in, v2.Original)
		}
	}
}

// Invalid versions: ParseNoCanon should also return ok=false.
func TestParseNoCanon_Invalid(t *testing.T) {
	for _, tt := range tests {
		if tt.out != "" {
			continue
		}
		_, ok := ParseNoCanon(tt.in)
		if ok {
			t.Fatalf("ParseNoCanon(%q) must be invalid (ok=false)", tt.in)
		}
	}
}

// Compare consistency: comparison after Parse and ParseNoCanon is the same.
func TestParseNoCanon_CompareConsistency(t *testing.T) {
	// возьмём только валидные входы
	vals := make([]string, 0, len(tests))
	for _, tt := range tests {
		if tt.out != "" {
			vals = append(vals, tt.in)
		}
	}
	for i := range vals {
		for j := range vals {
			vp1, _ := Parse(vals[i])
			vp2, _ := Parse(vals[j])
			vn1, _ := ParseNoCanon(vals[i])
			vn2, _ := ParseNoCanon(vals[j])

			if gotP, gotN := vp1.Compare(vp2), vn1.Compare(vn2); gotP != gotN {
				t.Fatalf("Compare mismatch %q vs %q: Parse=%d ParseNoCanon=%d",
					vals[i], vals[j], gotP, gotN)
			}
		}
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

var sinkInt int // чтобы so that the compiler does not throw out calls

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

func BenchmarkParseNoCanon(b *testing.B) {
	b.ReportAllocs()
	n := 0
	for i := 0; i < b.N; i++ {
		for _, s := range benchInputs {
			v, ok := ParseNoCanon(s)
			if ok {
				n += v.Major
			}
		}
	}
	sinkInt = n
}
