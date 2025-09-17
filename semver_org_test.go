// semver_semverorg_test.go
package semver

import (
	"regexp"
	"strings"
	"testing"
)

// reSemver is semver.org regex with optional leading v/V.
var reSemver = regexp.MustCompile(`^[vV]?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Valid list from semver.org (without leading v).
var semverOrgValid = []string{
	"0.0.4", "1.2.3", "10.20.30",
	"1.1.2-prerelease+meta", "1.1.2+meta", "1.1.2+meta-valid",
	"1.0.0-alpha", "1.0.0-beta", "1.0.0-alpha.beta", "1.0.0-alpha.beta.1",
	"1.0.0-alpha.1", "1.0.0-alpha0.valid", "1.0.0-alpha.0valid",
	"1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
	"1.0.0-rc.1+build.1", "2.0.0-rc.1+build.123", "1.2.3-beta",
	"10.2.3-DEV-SNAPSHOT", "1.2.3-SNAPSHOT-123", "1.0.0", "2.0.0", "1.1.7",
	"2.0.0+build.1848", "2.0.1-alpha.1227", "1.0.0-alpha+beta",
	"1.2.3----RC-SNAPSHOT.12.9.1--.12+788",
	"1.2.3----R-S.12.9.1--.12+meta",
	"1.2.3----RC-SNAPSHOT.12.9.1--.12",
	"1.0.0+0.build.1-rc.10000aaa-kk-0.1",
	"9999999999999999.99999999999.9999999999",
	"1.0.0-0A.is.legal",
	"1.5.0+20130313144700", "1.5.0-rc.0+X-TEST", "1.5.0-rc.0+build.1",
	"1.5.0-rc.0+20130313144700", "1.5.0-pre-zz",
	"1.5.0-beta2", "1.5.0-beta.2+20130313144700",
	"1.5.0-beta.2+build.1", "1.5.0-beta.2+X-TEST",
	"1.5.0-alpha.0", "1.5.0-alpha+X-TEST", "1.5.0-alpha+20130313144700",
	"1.5.0-alpha+build.1", "1.5.0-alpha",
	"1.5.0-1+X-TEST", "1.5.0-1+build.1", "1.5.0-1+20130313144700", "1.5.0-1",
	"1.2.3--alpha", "1.2.3--", "1.2.2+meta-pre.sha.256a",
}

// Invalid list from semver.org (without leading v).
var semverOrgInvalid = []string{
	"1", "1.2",
	"1.2.3-0123", "1.2.3-0123.0123", "1.1.2+.123",
	"+invalid", "-invalid", "-invalid+invalid", "-invalid.01",
	"alpha", "alpha.beta", "alpha.beta.1", "alpha.1",
	"alpha+beta", "alpha_beta", "alpha.", "alpha..", "beta",
	"1.0.0-alpha_beta", "-alpha.", "1.0.0-alpha..", "1.0.0-alpha..1",
	"1.0.0-alpha...1", "1.0.0-alpha....1", "1.0.0-alpha.....1",
	"1.0.0-alpha......1", "1.0.0-alpha.......1",
	"01.1.1", "1.01.1", "1.1.01",
	"1.2.3.DEV", "1.2-SNAPSHOT",
	"1.2.31.2.3----RC-SNAPSHOT.12.09.1--..12+788",
	"1.2-RC-SNAPSHOT", "-1.0.3-gamma+b7718", "+justmeta",
	"9.8.7+meta+meta", "9.8.7-whatever+meta+meta",
	"9999999999999999.99999999999.9999999999----RC-SNAPSHOT.12.09.1--------------------------------..12",
}

// isShorthand reports if s is MAJOR or MAJOR.MINOR (your extension).
var reShorthand = regexp.MustCompile(`^[vV]?\d+(?:\.\d+)?$`)

// canonFromRegex builds expected canonical "vMAJOR.MINOR.PATCH[-PRERELEASE]"
// from reSemver capture groups (build is intentionally stripped).
func canonFromRegex(s string) (string, bool) {
	m := reSemver.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}

	// m[1]=major, m[2]=minor, m[3]=patch, m[4]=pre, m[5]=build
	res := "v" + m[1] + "." + m[2] + "." + m[3]
	if m[4] != "" {
		res += "-" + m[4]
	}

	return res, true
}

// makeWithV returns s and s with a leading 'v' (if not already).
func makeWithV(s string) (string, string) {
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		return s, s
	}

	return s, "v" + s
}

// Valid cases: regex-valid => our Parse must be valid, Canonical must match regex-based canonical.
func TestSemverOrg_Valid_Parity(t *testing.T) {
	for _, s := range semverOrgValid {
		want, ok := canonFromRegex(s)
		if !ok {
			t.Fatalf("regex didn't match valid case %q", s)
		}

		a, b := makeWithV(s)
		inputs := []string{a}
		if b != a { // dedup when s already has v/V
			inputs = append(inputs, b)
		}

		for _, input := range inputs {
			v, ok := Parse(input)
			if !ok || !v.Valid {
				t.Fatalf("Parse(%q) invalid, want valid", input)
			}

			if got := v.Canonical(); got != want {
				t.Errorf("Canonical(%q)=%q, want %q", input, got, want)
			}
		}
	}
}

// Invalid cases: regex-invalid => our Parse must be invalid,
// except shorthand MAJOR / MAJOR.MINOR, which are valid in our dialect.
func TestSemverOrg_Invalid_Parity_WithShorthand(t *testing.T) {
	for _, s := range semverOrgInvalid {
		if reShorthand.MatchString(s) {
			// Your extension: treat as valid and normalize to vX.0.0 / vX.Y.0
			v, ok := Parse(s)
			if !ok || !v.Valid {
				t.Fatalf("Parse(%q) invalid, want valid shorthand", s)
			}

			parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
			want := "v" + parts[0]
			if len(parts) == 1 {
				want += ".0.0"
			} else {
				want += "." + parts[1] + ".0"
			}

			if got := v.Canonical(); got != want {
				t.Errorf("Canonical(%q)=%q, want %q", s, got, want)
			}

			continue
		}

		// Otherwise must be invalid for us too.
		if v, ok := Parse(s); ok || v.Valid {
			t.Errorf("Parse(%q) valid, want invalid", s)
		}
	}
}
