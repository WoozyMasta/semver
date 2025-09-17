package semver

import "strconv"

// Parse parses a version string into Semver and eagerly builds Canonical (1 alloc on success).
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
	cur := cursor{preStart: -1, preEnd: -1, buildStart: -1, buildEnd: -1}
	var pre, build string

	if i < len(raw) && raw[i] == '-' {
		ps, pe, next, ok := parsePrerelease(raw, i+1)
		if !ok {
			return Semver{Original: orig, Valid: false}, false
		}
		cur.preStart, cur.preEnd = ps, pe
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
		cur.buildStart, cur.buildEnd = bs, be
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
		cursor:     cur,
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

	// quick length guard: >19 digits can't fit into 64-bit anyway
	if j-i > 19 {
		return 0, i, false
	}

	n, _ := strconv.Atoi(raw[i:j]) // safe for small segments
	return n, j, true
}

// parsePrerelease validates prerelease and returns bounds within raw.
// 'start' is index right after '-'. Returns (preStart, preEnd, nextIndex, ok).
func parsePrerelease(raw string, start int) (int, int, int, bool) {
	i := start
	partStart := start
	for i < len(raw) && raw[i] != '+' {
		c := raw[i]
		if !(isIdentChar(c) || c == '.') {
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
		if !(isIdentChar(c) || c == '.') {
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

// isNum reports whether v consists entirely of digits.
func isNum(v string) bool {
	i := 0
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}

	return i == len(v)
}

// comparePrerelease compares two prerelease identifiers a and b
// according to Semantic Versioning rules.
//
// Rules:
//   - Empty string (no prerelease) has higher precedence.
//   - Identifiers are compared dot by dot.
//   - Numeric identifiers are compared numerically.
//   - Non-numeric identifiers are compared lexically in ASCII order.
//   - Numeric identifiers have lower precedence than non-numeric.
func comparePrerelease(a, b string) int {
	// Equal?
	if a == b {
		return 0
	}

	// Empty (release) is higher precedence than any pre-release.
	if a == "" {
		return +1
	}
	if b == "" {
		return -1
	}

	// Work with "-a" and "-b" to reuse the original state machine.
	x := "-" + a
	y := "-" + b
	for x != "" && y != "" {
		x = x[1:] // skip - or .
		y = y[1:]
		var dx, dy string
		dx, x = nextIdent(x)
		dy, y = nextIdent(y)
		if dx != dy {
			ix := isNum(dx)
			iy := isNum(dy)
			if ix != iy {
				if ix {
					return -1
				}

				return +1
			}

			if ix {
				// numeric: compare by length then lexicographically
				if len(dx) < len(dy) {
					return -1
				}
				if len(dx) > len(dy) {
					return +1
				}
			}

			if dx < dy {
				return -1
			}
			return +1
		}
	}

	if x == "" {
		return -1
	}

	return +1
}

// nextIdent returns the next identifier in x (up to '.'), and the rest.
func nextIdent(x string) (dx, rest string) {
	i := 0
	for i < len(x) && x[i] != '.' {
		i++
	}

	if i >= len(x) {
		return x, ""
	}

	return x[:i], x[i:]
}
