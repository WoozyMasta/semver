package semver

import "strings"

type PrintFlags uint16

const (
	PrintPrefixV PrintFlags = 1 << iota // always 'v'
	PrintPrefixNoV

	// components
	PrintMajor
	PrintMinor
	PrintPatch

	// include prerelease/build
	PrintPrerelease
	PrintBuild

	// MAJOR.MINOR.PATCH
	PrintMaskRelease = PrintMajor | PrintMinor | PrintPatch

	// vMAJOR.MINOR.PATCH[-PRERELEASE]
	PrintMaskCanonical = PrintPrefixV | PrintMaskRelease | PrintPrerelease

	// MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
	PrintMaskSemVer = PrintPrefixNoV | PrintMaskRelease | PrintPrerelease | PrintBuild

	// Preserve original prefix style and print everything available.
	PrintMaskDefault = PrintMaskRelease | PrintPrerelease | PrintBuild
)

// Print renders according to mask. It never invents prerelease/build, but
// zero-fills absent MINOR/PATCH to keep semver shape if they are requested.
func (v *Semver) Print(mask PrintFlags) string {
	if !v.Valid {
		return ""
	}

	// decide prefix
	var pfx byte
	switch {
	case (mask & PrintPrefixV) != 0:
		pfx = 'v'
	case (mask & PrintPrefixNoV) != 0:
		pfx = 0
	default:
		if v.HasV() && len(v.Original) > 0 {
			pfx = v.Original[0] // preserve exact 'v' or 'V'
		}
	}

	// determine which release parts are requested
	reqMajor := (mask & PrintMajor) != 0
	reqMinor := (mask & PrintMinor) != 0
	reqPatch := (mask & PrintPatch) != 0

	// zero-filled values if absent in input
	maj := v.Major // major is always parsed for valid semver
	min := v.Minor
	pat := v.Patch
	if reqMinor && (v.Flags&FlagHasMinor) == 0 {
		min = 0
	}
	if reqPatch && (v.Flags&FlagHasPatch) == 0 {
		pat = 0
	}

	// semver shape guard: if PATCH is requested but MINOR is not,
	// we must still print MINOR (zero-filled) to keep MAJOR.MINOR.PATCH.
	if reqPatch && !reqMinor {
		reqMinor = true
		if (v.Flags & FlagHasMinor) == 0 {
			min = 0
		}
	}
	// similarly, if MINOR is requested but MAJOR is not (weird), still print MAJOR to keep shape.
	if reqMinor && !reqMajor {
		reqMajor = true
	}

	// prerelease/build presence
	withPre := (mask&PrintPrerelease) != 0 && (v.Flags&FlagHasPre) != 0 && v.Prerelease != ""
	withBuild := (mask&PrintBuild) != 0 && (v.Flags&FlagHasBuild) != 0 && v.Build != ""

	// pre-calc length
	total := 0
	if pfx != 0 {
		total++
	}
	if reqMajor {
		total += digits10(maj)
	}
	if reqMinor {
		total += 1 + digits10(min)
	}
	if reqPatch {
		total += 1 + digits10(pat)
	}
	if withPre {
		total += 1 + len(v.Prerelease) // '-' + pre
	}
	if withBuild {
		total += 1 + len(v.Build) // '+' + build
	}
	if total == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(total)
	if pfx != 0 {
		b.WriteByte(pfx)
	}
	if reqMajor {
		writeInt(&b, maj)
	}
	if reqMinor {
		b.WriteByte('.')
		writeInt(&b, min)
	}
	if reqPatch {
		b.WriteByte('.')
		writeInt(&b, pat)
	}
	if withPre {
		b.WriteByte('-')
		b.WriteString(v.Prerelease)
	}
	if withBuild {
		b.WriteByte('+')
		b.WriteString(v.Build)
	}

	return b.String()
}

// Canonical returns "vMAJOR.MINOR.PATCH[-PRERELEASE]".
// Build metadata is intentionally stripped.
func (v *Semver) Canonical() string {
	return v.Print(PrintMaskCanonical)
}

// String implements fmt.Stringer.
// It renders "([v|V]?)MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]" preserving
// the original prefix style: if the original had 'v' or 'V', that exact
// rune is used, otherwise no prefix at all.
func (v *Semver) String() string {
	return v.Print(PrintMaskDefault)
}

// SemVer renders "MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]" with no prefix.
func (v *Semver) SemVer() string {
	return v.Print(PrintMaskSemVer)
}

// Full renders "([v|V]?)MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]".
// If preserve is true, it preserves the original prefix style:
//   - if Original started with 'v' or 'V' — uses that exact rune;
//   - otherwise — no prefix at all.
//
// If preserve is false, it always forces a lowercase 'v' prefix.
func (v *Semver) Full(preserve bool) string {
	mask := PrintMaskDefault
	if preserve {
		mask |= PrintPrefixV
	}

	// else preserve by leaving both prefix flags unset
	return v.Print(mask)
}

// MajorStr returns "vMAJOR". Empty if invalid.
// Always adds lowercase 'v' prefix.
func (v Semver) MajorStr() string {
	return (&v).Print(PrintPrefixV | PrintMajor)
}

// MajorMinorStr returns "vMAJOR.MINOR".
// Always adds lowercase 'v' prefix.
// Zero-fills MINOR if it was missing.
func (v Semver) MajorMinorStr() string {
	return (&v).Print(PrintPrefixV | PrintMajor | PrintMinor)
}

// ReleaseStr returns "vMAJOR.MINOR.PATCH".
// Always adds lowercase 'v' prefix.
// Zero-fills MINOR/PATCH if they were missing.
func (v Semver) ReleaseStr() string {
	return (&v).Print(PrintPrefixV | PrintMaskRelease)
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
