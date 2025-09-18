package semver

import (
	"strings"
)

// Semver represents a semantic version parsed from an input string.
// Prerelease and Build are zero-copy slices of Original when present.
// Major/Minor/Patch are normalized numeric values (no leading zeros).
// Flags expose what components were explicitly present in the input.
type Semver struct {
	// Original the raw input string (may be without "v")
	Original string

	// Prerelease optional pre-release part (no leading '-' in the value).
	// Zero-copy slice of Original when parsed; after mutators it may be a standalone string.
	Prerelease string

	// Build optional build metadata (no leading '+' in the value).
	// Zero-copy slice of Original when parsed; after mutators it may be a standalone string.
	Build string

	// Major numeric component (normalized, no leading zeros)
	Major int

	// Minor numeric component (normalized, no leading zeros)
	Minor int

	// Patch numeric component (normalized, no leading zeros)
	Patch int

	// Flags auxiliary flags affecting parsing or comparison behavior.
	Flags Flags

	// Valid indicates successful parsing
	Valid bool
}

// IsValid reports whether the Semver was successfully parsed.
func (v Semver) IsValid() bool {
	return v.Valid
}

// Compare compares v with w according to SemVer precedence.
// Returns -1 if v < w, 0 if v == w, +1 if v > w.
// Build metadata is ignored. Release (no prerelease) has higher precedence
// than any prerelease. Invalid versions are always smaller than valid ones.
func (v Semver) Compare(w Semver) int {
	if !v.Valid && !w.Valid {
		return 0
	}
	if !v.Valid {
		return -1
	}
	if !w.Valid {
		return 1
	}

	// numeric core
	if v.Major != w.Major {
		if v.Major < w.Major {
			return -1
		}
		return 1
	}

	if v.Minor != w.Minor {
		if v.Minor < w.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != w.Patch {
		if v.Patch < w.Patch {
			return -1
		}
		return 1
	}

	// prerelease: empty > any pre
	vHasPre := v.Flags&FlagHasPre != 0
	wHasPre := w.Flags&FlagHasPre != 0

	switch {
	case !vHasPre && !wHasPre:
		return 0
	case !vHasPre && wHasPre:
		return 1
	case vHasPre && !wHasPre:
		return -1
	default:
		return comparePrerelease(v.Prerelease, w.Prerelease)
	}
}

// Canonical returns "vMAJOR.MINOR.PATCH[-PRERELEASE]".
// Build metadata is intentionally stripped.
func (v *Semver) Canonical() string {
	if !v.Valid {
		return ""
	}

	preExtra := 0
	if v.Flags&FlagHasPre != 0 && v.Prerelease != "" {
		preExtra = 1 + len(v.Prerelease) // '-' + pre
	}
	total := 1 + digits10(v.Major) + 1 + digits10(v.Minor) + 1 + digits10(v.Patch) + preExtra

	var b strings.Builder
	b.Grow(total)
	b.WriteByte('v')
	writeInt(&b, v.Major)
	b.WriteByte('.')
	writeInt(&b, v.Minor)
	b.WriteByte('.')
	writeInt(&b, v.Patch)
	if preExtra > 0 {
		b.WriteByte('-')
		b.WriteString(v.Prerelease)
	}

	return b.String()
}

// String implements fmt.Stringer. It is identical to Canonical().
func (v *Semver) String() string {
	return v.Canonical()
}

// SemVer renders "MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]".
func (v *Semver) SemVer() string {
	if !v.Valid {
		return ""
	}

	preExtra := 0
	if v.Flags&FlagHasPre != 0 && v.Prerelease != "" {
		preExtra = 1 + len(v.Prerelease) // '-' + pre
	}
	preBuild := 0
	if v.Flags&FlagHasBuild != 0 && v.Build != "" {
		preBuild = 1 + len(v.Build) // '+' + build
	}

	total := digits10(v.Major) + 1 + digits10(v.Minor) + 1 + digits10(v.Patch) + preExtra + preBuild

	var b strings.Builder
	b.Grow(total)
	writeInt(&b, v.Major)
	b.WriteByte('.')
	writeInt(&b, v.Minor)
	b.WriteByte('.')
	writeInt(&b, v.Patch)

	if preExtra > 0 {
		b.WriteByte('-')
		b.WriteString(v.Prerelease)
	}
	if preBuild > 0 {
		b.WriteByte('+')
		b.WriteString(v.Build)
	}

	return b.String()
}

// Full renders "([v|V]?)MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]".
// If preserve is true, it always uses lowercase 'v' prefix.
// If preserve is false, it preserves the original prefix style:
//   - if Original started with 'v' or 'V' — uses that exact rune;
//   - otherwise — no prefix at all.
func (v *Semver) Full(preserve bool) string {
	if !v.Valid {
		return ""
	}

	// decide prefix
	var pfx byte
	if preserve {
		pfx = 'v'
	} else {
		if v.HasV() && len(v.Original) > 0 {
			pfx = v.Original[0]
		} else {
			pfx = 0 // no prefix
		}
	}

	preExtra := 0
	if v.Flags&FlagHasPre != 0 && v.Prerelease != "" {
		preExtra = 1 + len(v.Prerelease) // '-' + pre
	}
	preBuild := 0
	if v.Flags&FlagHasBuild != 0 && v.Build != "" {
		preBuild = 1 + len(v.Build) // '+' + build
	}

	total := int(0)
	if pfx != 0 {
		total++
	}
	total += 1 + digits10(v.Major) + 1 + digits10(v.Minor) + 1 + digits10(v.Patch) + preExtra + preBuild

	var b strings.Builder
	b.Grow(total)
	if pfx != 0 {
		b.WriteByte(pfx)
	}
	writeInt(&b, v.Major)
	b.WriteByte('.')
	writeInt(&b, v.Minor)
	b.WriteByte('.')
	writeInt(&b, v.Patch)

	if preExtra > 0 {
		b.WriteByte('-')
		b.WriteString(v.Prerelease)
	}
	if preBuild > 0 {
		b.WriteByte('+')
		b.WriteString(v.Build)
	}

	return b.String()
}

// MajorStr returns "vMAJOR". Empty if invalid.
func (v Semver) MajorStr() string {
	if v.Flags&FlagHasMajor == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(1 + digits10(v.Major))
	b.WriteByte('v')
	writeInt(&b, v.Major)

	return b.String()
}

// MajorMinorStr returns "vMAJOR.MINOR".
// Empty if invalid or MINOR wasn't present in the input.
func (v Semver) MajorMinorStr() string {
	if v.Flags&(FlagHasMajor|FlagHasMinor) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(1 + digits10(v.Major) + 1 + digits10(v.Minor))
	b.WriteByte('v')
	writeInt(&b, v.Major)
	b.WriteByte('.')
	writeInt(&b, v.Minor)

	return b.String()
}

// ReleaseStr returns "vMAJOR.MINOR.PATCH".
// Empty if invalid or PATCH wasn't present in the input.
func (v Semver) ReleaseStr() string {
	if v.Flags&(FlagHasMajor|FlagHasMinor|FlagHasPatch) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(1 + digits10(v.Major) + 1 + digits10(v.Minor) + 1 + digits10(v.Patch))
	b.WriteByte('v')
	writeInt(&b, v.Major)
	b.WriteByte('.')
	writeInt(&b, v.Minor)
	b.WriteByte('.')
	writeInt(&b, v.Patch)

	return b.String()
}

// Max returns the greater of two Semver values.
func (v Semver) Max(w Semver) Semver {
	if v.Compare(w) >= 0 {
		return v
	}

	return w
}

// IsGreater reports whether the receiver v represents a semantic version that is
// strictly greater than the provided version w according to the package's
// comparison rules. It returns true when v.Compare(w) > 0.
func (v Semver) IsGreater(w Semver) bool {
	return v.Compare(w) > 0
}

// IsLower reports whether the receiver v represents a semantic version that is
// strictly lower than the provided version w according to the package's
// comparison rules. It returns true when v.Compare(w) < 0.
func (v Semver) IsLower(w Semver) bool {
	return v.Compare(w) < 0
}

// IsEqual reports whether the receiver v is equal to the provided version w
// according to semantic version precedence rules. It returns true when v.Compare(w) == 0.
func (v Semver) IsEqual(w Semver) bool {
	return v.Compare(w) == 0
}

// writeInt writes a non-negative integer to the builder using a small stack buffer.
func writeInt(b *strings.Builder, x int) {
	// handle zero fast-path
	if x == 0 {
		b.WriteByte('0')
		return
	}

	var buf [20]byte // enough for int64
	i := len(buf)
	u := x
	for u > 0 {
		i--
		buf[i] = byte('0' + u%10)
		u /= 10
	}

	b.Write(buf[i:])
}

// digits10 returns number of decimal digits in a non-negative integer.
func digits10(x int) int {
	if x == 0 {
		return 1
	}

	n := 0
	for u := x; u > 0; u /= 10 {
		n++
	}

	return n
}
