/*
Package semver provides parsing and comparison of Semantic Versioning
strings (SemVer 2.0.0) with a few pragmatic deviations:

  - Leading "v"/"V" is optional on input (e.g. "1.2.3" or "v1.2.3").
  - Shorthands "MAJOR" and "MAJOR.MINOR" are accepted and normalized to
    "vMAJOR.0.0" and "vMAJOR.MINOR.0". Prerelease/build require full x.y.z.
  - Parsed value (Semver) exposes numeric fields (Major, Minor, Patch),
    prerelease and build metadata (zero-copy slices of Original on parse),
    and canonical string form via Canonical()/String().
  - Comparisons follow SemVer precedence and ignore build metadata.
    Invalid versions are considered smaller than valid ones.
  - Sorting via List uses SemVer precedence with a lexicographic
    tie-breaker on the original input to produce stable order.
  - Numeric components must fit into the host int size; overly large
    numbers are rejected as invalid.

Canonical form returned by Semver.Canonical() (and String()) always uses the
"v" prefix and strips build metadata: "vMAJOR.MINOR.PATCH[-PRERELEASE]".

# Quick examples

Parse and Canonical:

	v1, _ := Parse("1.2")
	v2, _ := Parse("v1.2.0-rc.1+build.5")
	_ = v1.Canonical() // "v1.2.0"
	_ = v2.Canonical() // "v1.2.0-rc.1"

Compare (release > prerelease on same core; build ignored):

	a, _ := Parse("1.2.4")
	b, _ := Parse("1.2.4-rc.1")
	c, _ := Parse("1.2.4+meta")
	_ = a.Compare(b)   // +1
	_ = a.Compare(c)   //  0

Bump* helpers (clear pre/build and normalize):

	x, _ := Parse("1.2.3-rc.1+meta")
	p, _ := x.BumpPatch() // "v1.2.4"
	m, _ := x.BumpMinor() // "v1.3.0"
	M, _ := x.BumpMajor() // "v2.0.0"
	_, _, _ = p.Canonical(), m.Canonical(), M.Canonical()

WithPre / WithBuild (validate segments per SemVer):

	y, _ := Parse("1")
	y1, _ := y.WithPre("alpha.1")     // Canonical: "v1.0.0-alpha.1"
	y2, _ := y1.WithBuild("meta.5")   // Full:      "v1.0.0-alpha.1+meta.5"
	_, _ = y1.Canonical(), y2.Full()

NextPrerelease (increments numeric tail; default base "rc"):

	z, _ := Parse("1.2.3-rc.9")
	z1, _ := z.NextPrerelease("rc")   // "v1.2.3-rc.10"
	z2, _ := Parse("1.2.3")           // no pre -> use base
	z2, _ = z2.NextPrerelease("")     // "v1.2.3-rc.1"
	_, _ = z1.Canonical(), z2.Canonical()

Full():

	u, _ := Parse("V1.2.3-rc.1+meta")
	_ = u.Full(true)  // "v1.2.3-rc.1+meta" (force 'v')
	_ = u.Full(false) // "V1.2.3-rc.1+meta" (preserve original 'V')
*/
package semver
