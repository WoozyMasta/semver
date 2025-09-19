package semver

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
