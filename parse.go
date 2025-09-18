package semver

// Parse parses a version string into Semver.
// It accepts an optional leading 'v'/'V' and the shorthand forms "MAJOR" and
// "MAJOR.MINOR" (which normalize to ".0.0" and ".0").
// Prerelease/build are only allowed when MAJOR.MINOR.PATCH are all present.
// Numeric components must fit into the host int size; otherwise the input
// is rejected as invalid.
func Parse(s string) (Semver, bool) {
	if s == "" {
		return Semver{Original: s, Valid: false}, false
	}
	orig := s
	flags := Flags(0)

	// Compute raw view (skip optional leading 'v'/'V') via offset, no extra field in struct.
	vOffset := 0
	if orig[0] == 'v' || orig[0] == 'V' {
		flags |= FlagHasV
		if len(orig) == 1 {
			return Semver{Original: orig, Valid: false}, false
		}
		vOffset = 1
	}
	raw := orig[vOffset:]

	i := 0

	// major (required)
	maj, n, ok := parseInt(raw, i)
	if !ok {
		return Semver{Original: orig, Valid: false}, false
	}
	flags |= FlagHasMajor
	i = n

	min, pat := 0, 0

	// minor (optional shorthand)
	if i < len(raw) && raw[i] == '.' {
		i++
		mm, n2, ok := parseInt(raw, i)
		if !ok {
			return Semver{Original: orig, Valid: false}, false
		}
		min = mm
		i = n2
		flags |= FlagHasMinor

		// patch (optional shorthand)
		if i < len(raw) && raw[i] == '.' {
			i++
			pp, n3, ok := parseInt(raw, i)
			if !ok {
				return Semver{Original: orig, Valid: false}, false
			}
			pat = pp
			i = n3
			flags |= FlagHasPatch
		}
	}

	if i < len(raw) && (raw[i] == '-' || raw[i] == '+') && flags&FlagHasPatch == 0 {
		return Semver{Original: orig, Valid: false}, false
	}

	// prerelease (optional, after '-')
	var pre, build string

	if i < len(raw) && raw[i] == '-' {
		ps, pe, next, ok := parsePrerelease(raw, i+1)
		if !ok {
			return Semver{Original: orig, Valid: false}, false
		}
		pre = orig[vOffset+ps : vOffset+pe] // zero-copy slice of Original
		i = next
		flags |= FlagHasPre
	}

	// build (optional, after '+')
	if i < len(raw) && raw[i] == '+' {
		bs, be, next, ok := parseBuild(raw, i+1)
		if !ok {
			return Semver{Original: orig, Valid: false}, false
		}
		build = orig[vOffset+bs : vOffset+be] // zero-copy slice of Original
		i = next
		flags |= FlagHasBuild
	}

	// nothing must remain
	if i != len(raw) {
		return Semver{Original: orig, Valid: false}, false
	}

	v := Semver{
		Original:   orig,
		Major:      maj,
		Minor:      min,
		Patch:      pat,
		Prerelease: pre,
		Build:      build,
		Flags:      flags,
		Valid:      true,
	}

	return v, true
}

// parseInt parses a non-negative int at raw[i:], SemVer rules (no leading zeros for multi-digit).
// Returns value, next index, ok.
func parseInt(raw string, i int) (val int, next int, ok bool) {
	// no digits
	if i >= len(raw) || raw[i] < '0' || raw[i] > '9' {
		return 0, i, false
	}

	// scan digits
	j := i + 1
	for j < len(raw) && raw[j] >= '0' && raw[j] <= '9' {
		j++
	}

	// reject leading zeros in multi-digit numbers
	if raw[i] == '0' && j-i > 1 {
		return 0, i, false
	}

	// accumulate with overflow check for host int
	const MaxInt = int(^uint(0) >> 1)
	n := 0
	for k := i; k < j; k++ {
		d := int(raw[k] - '0')
		if n > (MaxInt-d)/10 {
			return 0, i, false // overflow
		}
		n = n*10 + d
	}

	return n, j, true
}

// parsePrerelease validates prerelease and returns bounds within raw.
// 'start' is index right after '-'. Returns (preStart, preEnd, nextIndex, ok).
func parsePrerelease(raw string, start int) (int, int, int, bool) {
	i := start
	partStart := start
	for i < len(raw) && raw[i] != '+' {
		c := raw[i]
		if !isIdentChar(c) && c != '.' {
			return 0, 0, 0, false
		}

		if c == '.' {
			if partStart == i || isBadNum(raw[partStart:i]) {
				return 0, 0, 0, false
			}
			partStart = i + 1
		}
		i++
	}

	if partStart == i || isBadNum(raw[partStart:i]) {
		return 0, 0, 0, false
	}

	return start, i, i, true
}

// parseBuild validates build metadata and returns bounds within raw.
// 'start' is index after '+'. Returns (buildStart, buildEnd, nextIndex, ok).
func parseBuild(raw string, start int) (int, int, int, bool) {
	i := start
	partStart := start
	for i < len(raw) {
		c := raw[i]
		if !isIdentChar(c) && c != '.' {
			return 0, 0, 0, false
		}

		if c == '.' {
			if partStart == i {
				return 0, 0, 0, false
			}
			partStart = i + 1
		}
		i++
	}

	if partStart == i {
		return 0, 0, 0, false
	}

	return start, i, i, true
}

// isIdentChar reports whether c is a valid identifier character
// in prerelease/build metadata ([0-9A-Za-z-]).
func isIdentChar(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-'
}

// isBadNum reports whether v is a numeric identifier
// with leading zeroes (which are invalid).
func isBadNum(v string) bool {
	if v == "" {
		return false
	}

	i := 0
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}

	return i == len(v) && i > 1 && v[0] == '0'
}
