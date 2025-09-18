package semver

import "strings"

// BumpPatch returns v with Patch+1 and clears prerelease/build.
// Returns (zero, false) if v is invalid.
func (v Semver) BumpPatch() (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	nv := v
	nv.Patch++
	nv.Prerelease, nv.Build = "", ""
	nv.Flags |= FlagHasMajor | FlagHasMinor | FlagHasPatch
	nv.Flags &^= (FlagHasPre | FlagHasBuild)
	nv.Original = nv.Full(false)

	return nv, true
}

// BumpMinor returns v with Minor+1, Patch=0 and clears prerelease/build.
func (v Semver) BumpMinor() (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	nv := v
	nv.Minor++
	nv.Patch = 0
	nv.Prerelease, nv.Build = "", ""
	nv.Flags |= FlagHasMajor | FlagHasMinor | FlagHasPatch
	nv.Flags &^= (FlagHasPre | FlagHasBuild)
	nv.Original = nv.Full(false)

	return nv, true
}

// BumpMajor returns v with Major+1, Minor=0, Patch=0 and clears prerelease/build.
func (v Semver) BumpMajor() (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	nv := v
	nv.Major++
	nv.Minor, nv.Patch = 0, 0
	nv.Prerelease, nv.Build = "", ""
	nv.Flags |= FlagHasMajor | FlagHasMinor | FlagHasPatch
	nv.Flags &^= (FlagHasPre | FlagHasBuild)
	nv.Original = nv.Full(false)

	return nv, true
}

// WithPre returns v with given prerelease (without leading '-'). Validates per SemVer.
// If v was a shorthand (no MINOR/PATCH), they are normalized to 0.
// Returns (zero, false) if v is invalid or prerelease is invalid.
func (v Semver) WithPre(pre string) (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	// validate prerelease using package parser
	if pre != "" {
		raw := "-" + pre
		if _, _, next, ok := parsePrerelease(raw, 1); !ok || next != len(raw) {
			return Semver{Original: v.Original, Valid: false}, false
		}
	}

	nv := v
	if nv.Flags&FlagHasMinor == 0 {
		nv.Minor = 0
		nv.Flags |= FlagHasMinor
	}

	if nv.Flags&FlagHasPatch == 0 {
		nv.Patch = 0
		nv.Flags |= FlagHasPatch
	}

	nv.Prerelease = pre
	if pre != "" {
		nv.Flags |= FlagHasPre
	} else {
		nv.Flags &^= FlagHasPre
	}

	nv.Original = nv.Full(false)

	return nv, true
}

// WithBuild returns v with given build metadata (without leading '+'). Validates per SemVer.
// If v was a shorthand (no MINOR/PATCH), they are normalized to 0.
// Returns (zero, false) if v is invalid or build is invalid.
func (v Semver) WithBuild(build string) (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	if build != "" {
		raw := "+" + build
		if _, _, next, ok := parseBuild(raw, 1); !ok || next != len(raw) {
			return Semver{Original: v.Original, Valid: false}, false
		}
	}

	nv := v
	if nv.Flags&FlagHasMinor == 0 {
		nv.Minor = 0
		nv.Flags |= FlagHasMinor
	}
	if nv.Flags&FlagHasPatch == 0 {
		nv.Patch = 0
		nv.Flags |= FlagHasPatch
	}

	nv.Build = build
	if build != "" {
		nv.Flags |= FlagHasBuild
	} else {
		nv.Flags &^= FlagHasBuild
	}

	nv.Original = nv.Full(false)

	return nv, true
}

// StripPre removes prerelease if present.
func (v Semver) StripPre() (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	nv := v
	nv.Prerelease = ""
	nv.Flags &^= FlagHasPre
	nv.Original = nv.Full(false)

	return nv, true
}

// StripBuild removes build metadata if present.
func (v Semver) StripBuild() (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	nv := v
	nv.Build = ""
	nv.Flags &^= FlagHasBuild
	nv.Original = nv.Full(false)

	return nv, true
}

// NextPrerelease increments the last numeric identifier.
// If none, appends ".1". If prerelease empty, sets to base (e.g. "rc.1").
// base is used only when current prerelease is empty; pass "" to default "rc".
func (v Semver) NextPrerelease(base string) (Semver, bool) {
	if !v.Valid {
		return Semver{Original: v.Original, Valid: false}, false
	}

	nv := v
	cur := nv.Prerelease

	if cur == "" {
		if base == "" {
			base = "rc"
		}

		nv.Prerelease = base + ".1"
		nv.Flags |= FlagHasPre

		return nv, true
	}

	parts := strings.Split(cur, ".")
	last := parts[len(parts)-1]
	if isNum(last) {
		// increment numeric tail
		b := []byte(last)
		carry := 1
		for i := len(b) - 1; i >= 0 && carry == 1; i-- {
			if b[i] == '9' {
				b[i] = '0'
			} else {
				b[i]++
				carry = 0
			}
		}

		if carry == 1 {
			b = append([]byte{'1'}, b...)
		}

		parts[len(parts)-1] = string(b)
	} else {
		parts = append(parts, "1")
	}

	nv.Prerelease = strings.Join(parts, ".")
	nv.Flags |= FlagHasPre
	nv.Original = nv.Full(false)

	return nv, true
}
