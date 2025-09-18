package semver

import (
	"testing"
)

func TestBumpCore(t *testing.T) {
	v, _ := Parse("1.2.3-rc.1+build.5")

	vp, ok := v.BumpPatch()
	if !ok || vp.Canonical() != "v1.2.4" || vp.HasPre() || vp.HasBuild() {
		t.Fatalf("BumpPatch: got %q (pre=%v build=%v)", vp.Canonical(), vp.HasPre(), vp.HasBuild())
	}

	vm, ok := v.BumpMinor()
	if !ok || vm.Canonical() != "v1.3.0" || vm.HasPre() || vm.HasBuild() {
		t.Fatalf("BumpMinor: got %q", vm.Canonical())
	}

	vM, ok := v.BumpMajor()
	if !ok || vM.Canonical() != "v2.0.0" || vM.HasPre() || vM.HasBuild() {
		t.Fatalf("BumpMajor: got %q", vM.Canonical())
	}
}

func TestWithPreAndBuild(t *testing.T) {
	v, _ := Parse("1") // shorthand

	v1, ok := v.WithPre("alpha.1")
	if !ok || v1.Canonical() != "v1.0.0-alpha.1" {
		t.Fatalf("WithPre: got %q", v1.Canonical())
	}
	v2, ok := v1.WithBuild("meta.5")
	if !ok || v2.Full(true) != "v1.0.0-alpha.1+meta.5" {
		t.Fatalf("WithBuild: got %q", v2.Full(true))
	}

	// invalid pre/build
	if _, ok := v.WithPre("01"); ok { // leading zero numeric
		t.Fatalf("WithPre accepted invalid")
	}
	if _, ok := v.WithBuild("bad..seg"); ok {
		t.Fatalf("WithBuild accepted invalid")
	}
}

func TestCompareHelpers(t *testing.T) {
	// release > prerelease when core equal
	a, _ := Parse("1.2.4")
	b, _ := Parse("1.2.4-rc.1")
	if !a.IsGreater(b) || b.IsGreater(a) || a.IsEqual(b) {
		t.Fatalf("compare helpers broken (release vs prerelease on same core)")
	}

	// strictly greater core
	c, _ := Parse("1.2.3")
	if !a.IsGreater(c) || c.IsGreater(a) || a.IsEqual(c) {
		t.Fatalf("compare helpers broken (different cores)")
	}

	// equal (build ignored)
	d1, _ := Parse("1.2.3")
	d2, _ := Parse("1.2.3+meta")
	if !d1.IsEqual(d2) || d1.IsGreater(d2) || d2.IsGreater(d1) {
		t.Fatalf("compare helpers broken (equality)")
	}
}

func TestStripPre(t *testing.T) {
	v, _ := Parse("1.2.3-rc.1+meta")

	nv, ok := v.StripPre()
	if !ok {
		t.Fatalf("StripPre returned !ok")
	}

	// prerelease removed, build preserved
	if nv.Canonical() != "v1.2.3" {
		t.Fatalf("Canonical after StripPre = %q, want v1.2.3", nv.Canonical())
	}
	if got := nv.Original; got != "1.2.3+meta" {
		t.Fatalf("Full after StripPre = %q, want 1.2.3+meta", got)
	}

	if nv.HasPre() {
		t.Fatalf("HasPre after StripPre = true, want false")
	}
	if !nv.HasBuild() {
		t.Fatalf("HasBuild after StripPre = false, want true")
	}

	nv2, ok := nv.StripPre()
	if !ok || nv2.Canonical() != "v1.2.3" || nv2.HasPre() {
		t.Fatalf("StripPre idempotent failed: Canonical=%q HasPre=%v ok=%v",
			nv2.Canonical(), nv2.HasPre(), ok)
	}
}

func TestStripBuild(t *testing.T) {
	v, _ := Parse("1.2.3-rc.1+meta")

	nv, ok := v.StripBuild()
	if !ok {
		t.Fatalf("StripBuild returned !ok")
	}

	// build removed, prerelease preserved
	if nv.Canonical() != "v1.2.3-rc.1" {
		t.Fatalf("Canonical after StripBuild = %q, want v1.2.3-rc.1", nv.Canonical())
	}
	if got := nv.Original; got != "1.2.3-rc.1" {
		t.Fatalf("Full after StripBuild = %q, want 1.2.3-rc.1", got)
	}

	if nv.HasBuild() {
		t.Fatalf("HasBuild after StripBuild = true, want false")
	}
	if !nv.HasPre() {
		t.Fatalf("HasPre after StripBuild = false, want true")
	}
}

func TestStrip_InvalidAndShorthand(t *testing.T) {
	// invalid inputs → ok=false
	for _, in := range []string{"v1-pre", "1.2-pre", "bad"} {
		v, _ := Parse(in)
		if _, ok := v.StripPre(); ok {
			t.Fatalf("StripPre accepted invalid %q", in)
		}
		if _, ok := v.StripBuild(); ok {
			t.Fatalf("StripBuild accepted invalid %q", in)
		}
	}

	// shorthands (no pre/build) → ok=true, no change
	for _, in := range []string{"1", "1.2", "1.2.3"} {
		v, _ := Parse(in)
		v1, ok := v.StripPre()
		if !ok || v1.Canonical() != v.Canonical() || v1.HasPre() {
			t.Fatalf("StripPre shorthand %q: Canonical=%q HasPre=%v ok=%v",
				in, v1.Canonical(), v1.HasPre(), ok)
		}
		v2, ok := v.StripBuild()
		if !ok || v2.Canonical() != v.Canonical() || v2.HasBuild() {
			t.Fatalf("StripBuild shorthand %q: Canonical=%q HasBuild=%v ok=%v",
				in, v2.Canonical(), v2.HasBuild(), ok)
		}
	}

	// pure-build only
	vb, _ := Parse("1.2.3+meta")
	vb1, ok := vb.StripBuild()
	if !ok || vb1.Full(true) != "v1.2.3" || vb1.HasBuild() {
		t.Fatalf("StripBuild on build-only: Full=%q HasBuild=%v ok=%v",
			vb1.Full(true), vb1.HasBuild(), ok)
	}

	// pure-pre only
	vp, _ := Parse("1.2.3-rc.1")
	vp1, ok := vp.StripPre()
	if !ok || vp1.Canonical() != "v1.2.3" || vp1.HasPre() {
		t.Fatalf("StripPre on pre-only: Canonical=%q HasPre=%v ok=%v",
			vp1.Canonical(), vp1.HasPre(), ok)
	}
}

func TestStrip_OrderBothWays(t *testing.T) {
	v, _ := Parse("1.2.3-rc.1+meta")

	// build then pre
	a, ok := v.StripBuild()
	if !ok {
		t.Fatalf("StripBuild failed")
	}
	a, ok = a.StripPre()
	if !ok {
		t.Fatalf("StripPre failed after StripBuild")
	}
	if a.Canonical() != "v1.2.3" || a.Full(true) != "v1.2.3" || !a.IsRelease() {
		t.Fatalf("after StripBuild->StripPre: Canonical=%q Full=%q IsRelease=%v",
			a.Canonical(), a.Full(true), a.IsRelease())
	}

	// pre then build
	v2, _ := Parse("1.2.3-rc.1+meta")
	b, ok := v2.StripPre()
	if !ok {
		t.Fatalf("StripPre failed")
	}
	b, ok = b.StripBuild()
	if !ok {
		t.Fatalf("StripBuild failed after StripPre")
	}
	if b.Canonical() != "v1.2.3" || b.Full(true) != "v1.2.3" || !b.IsRelease() {
		t.Fatalf("after StripPre->StripBuild: Canonical=%q Full=%q IsRelease=%v",
			b.Canonical(), b.Full(true), b.IsRelease())
	}
}

func TestNextPrerelease(t *testing.T) {
	tests := []struct {
		in, base, out string
	}{
		{"1.2.3", "rc", "v1.2.3-rc.1"},
		{"1.2.3-rc.1", "rc", "v1.2.3-rc.2"},
		{"1.2.3-rc.9", "rc", "v1.2.3-rc.10"},
		{"1.2.3-rc.9.beta", "rc", "v1.2.3-rc.9.beta.1"},
		{"1.2.3-rc.9.beta.5", "rc", "v1.2.3-rc.9.beta.6"},
		{"1.2.3-alpha", "rc", "v1.2.3-alpha.1"},
		{"1.2.3-alpha.beta", "rc", "v1.2.3-alpha.beta.1"},
		{"1.2.3-alpha.beta.5", "rc", "v1.2.3-alpha.beta.6"},
		{"1.2.3-9", "rc", "v1.2.3-10"},
		{"1.2.3-0", "rc", "v1.2.3-1"},
		{"1.2.3-a.b.c", "rc", "v1.2.3-a.b.c.1"},
		{"1.2.3-a.b.9", "rc", "v1.2.3-a.b.10"},
		{"1.2.3-rc.1+build.5", "rc", "v1.2.3-rc.2"}, // build ignored
		{"1.2.3+build.5", "rc", "v1.2.3-rc.1"},      // build ignored
	}

	for _, tt := range tests {
		v, _ := Parse(tt.in)
		vn, ok := v.NextPrerelease(tt.base)
		if !ok || vn.Canonical() != tt.out {
			t.Errorf("NextPrerelease(%q, %q) = %q, %v; want %q, true", tt.in, tt.base, vn.Canonical(), ok, tt.out)
		}
	}

	// invalid input
	vBad := Semver{Original: "bad", Valid: false}
	if _, ok := vBad.NextPrerelease("rc"); ok {
		t.Errorf("NextPrerelease accepted invalid input")
	}

	// invalid base
	v, _ := Parse("1.2.3")
	vn, ok := v.NextPrerelease("") // default base
	if !ok || vn.Canonical() != "v1.2.3-rc.1" {
		t.Errorf("NextPrerelease default base: got %q, ok=%v; want v1.2.3-rc.1, true", vn.Canonical(), ok)
	}
}
