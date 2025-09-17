package semver

import "strconv"

// Semver represents a semantic version parsed from an input string.
type Semver struct {
	// Original the raw input string (may be without "v")
	Original string

	// Prerelease optional pre-release part (without leading '-')
	// Zero-copy slice of Original.
	Prerelease string

	// Build optional build metadata (without leading '+')
	// Zero-copy slice of Original.
	Build string

	// cursor caches bounds of pre/build within raw = Original[vOffset:].
	// The bounds do NOT include the leading '-' or '+'.
	cursor cursor

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

type cursor struct {
	preStart, preEnd     int // bounds in raw (no '-' in slice)
	buildStart, buildEnd int // bounds in raw (no '+')
}

// IsValid reports whether the Semver was successfully parsed.
func (v Semver) IsValid() bool { return v.Valid }

// Compare compares v with w according to semantic version precedence.
// Returns -1 if v < w, 0 if v == w, +1 if v > w.
// Invalid versions are always considered smaller than valid ones.
// Build metadata is ignored for ordering.
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
func (v *Semver) Canonical() string {
	if !v.Valid {
		return ""
	}

	b := make([]byte, 0, 16+len(v.Prerelease))
	b = append(b, 'v')
	b = strconv.AppendInt(b, int64(v.Major), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Minor), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Patch), 10)
	if v.Flags&FlagHasPre != 0 {
		b = append(b, '-')
		b = append(b, v.Prerelease...)
	}

	return string(b)
}

// String returns "vMAJOR.MINOR.PATCH[-PRERELEASE]".
func (v *Semver) String() string {
	return v.Canonical()
}

// Full returns "vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]" for display/logging.
func (v *Semver) Full() string {
	if !v.Valid {
		return ""
	}

	b := make([]byte, 0, 16+len(v.Prerelease)+1+len(v.Build))
	b = append(b, 'v')
	b = strconv.AppendInt(b, int64(v.Major), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Minor), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Patch), 10)
	if v.Flags&FlagHasPre != 0 {
		b = append(b, '-')
		b = append(b, v.Prerelease...)
	}
	if v.Flags&FlagHasBuild != 0 {
		b = append(b, '+')
		b = append(b, v.Build...)
	}

	return string(b)
}

// MajorStr returns "vMAJOR". Empty if invalid.
func (v Semver) MajorStr() string {
	if v.Flags&FlagHasMajor == 0 {
		return ""
	}

	b := make([]byte, 0, 8)
	b = append(b, 'v')
	b = strconv.AppendInt(b, int64(v.Major), 10)

	return string(b)
}

// MajorMinorStr returns "vMAJOR.MINOR". Empty if invalid.
func (v Semver) MajorMinorStr() string {
	if v.Flags&(FlagHasMajor|FlagHasMinor) == 0 {
		return ""
	}

	b := make([]byte, 0, 12)
	b = append(b, 'v')
	b = strconv.AppendInt(b, int64(v.Major), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Minor), 10)

	return string(b)
}

// ReleaseStr returns "vMAJOR.MINOR.PATCH". Empty if invalid.
func (v Semver) ReleaseStr() string {
	if v.Flags&(FlagHasMajor|FlagHasMinor|FlagHasPatch) == 0 {
		return ""
	}

	b := make([]byte, 0, 16)
	b = append(b, 'v')
	b = strconv.AppendInt(b, int64(v.Major), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Minor), 10)
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(v.Patch), 10)

	return string(b)
}

// Max returns the greater of two Semver values.
func (v Semver) Max(w Semver) Semver {
	if v.Compare(w) >= 0 {
		return v
	}

	return w
}
