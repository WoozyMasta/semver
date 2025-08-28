package semver

// parsed is the internal representation of a semantic version split
// into its components. Used only inside this package.
type parsed struct {
	Major      string // MAJOR number
	Minor      string // MINOR number
	Patch      string // PATCH number
	Prerelease string // prerelease part, including leading '-' or ""
	Build      string // build metadata, including leading '+' or ""
}

// parseInternal parses v into a parsed struct.
//
// Differences from x/mod/semver:
//   - A leading "v" or "V" is optional.
//   - Shorthands are supported: "MAJOR" or "MAJOR.MINOR"
//     are expanded to "MAJOR.0.0" or "MAJOR.MINOR.0".
//
// It returns ok=false if the string is not a valid semantic version.
func parseInternal(v string) (p parsed, ok bool) {
	if v == "" {
		return
	}

	// make 'v' optional
	if v[0] == 'v' || v[0] == 'V' {
		v = v[1:]
	}

	// major
	p.Major, v, ok = parseInt(v)
	if !ok {
		return
	}

	// minor
	if v == "" {
		p.Minor = "0"
		p.Patch = "0"
		return
	}
	if v[0] != '.' {
		ok = false
		return
	}

	p.Minor, v, ok = parseInt(v[1:])
	if !ok {
		return
	}

	// patch
	if v == "" {
		p.Patch = "0"
		return
	}
	if v[0] != '.' {
		ok = false
		return
	}
	p.Patch, v, ok = parseInt(v[1:])
	if !ok {
		return
	}

	// prerelease
	if len(v) > 0 && v[0] == '-' {
		var t string
		t, v, ok = parsePrerelease(v)
		if !ok {
			return
		}
		p.Prerelease = t
	}

	// build
	if len(v) > 0 && v[0] == '+' {
		var t string
		t, v, ok = parseBuild(v)
		if !ok {
			return
		}
		p.Build = t
	}

	// leftovers invalid
	if v != "" {
		ok = false
		return
	}
	ok = true
	return
}

// parseInt parses a non-negative integer prefix from v.
// Returns the digits, the remainder, and ok.
// Multi-digit numbers must not start with '0'.
func parseInt(v string) (t, rest string, ok bool) {
	if v == "" {
		return
	}

	if v[0] < '0' || '9' < v[0] {
		return
	}

	i := 1
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}

	// reject leading zeros in multi-digit
	if v[0] == '0' && i != 1 {
		return
	}

	return v[:i], v[i:], true
}

// parsePrerelease parses a prerelease part (starting with '-').
// Returns the part including leading '-', the remainder, and ok.
func parsePrerelease(v string) (t, rest string, ok bool) {
	if v == "" || v[0] != '-' {
		return
	}

	i := 1
	start := 1
	for i < len(v) && v[i] != '+' {
		if !isIdentChar(v[i]) && v[i] != '.' {
			return
		}

		if v[i] == '.' {
			if start == i || isBadNum(v[start:i]) {
				return
			}
			start = i + 1
		}
		i++
	}

	if start == i || isBadNum(v[start:i]) {
		return
	}

	return v[:i], v[i:], true
}

// parseBuild parses a build metadata part (starting with '+').
// Returns the part including leading '+', the remainder, and ok.
func parseBuild(v string) (t, rest string, ok bool) {
	if v == "" || v[0] != '+' {
		return
	}

	i := 1
	start := 1
	for i < len(v) {
		if !isIdentChar(v[i]) && v[i] != '.' {
			return
		}

		if v[i] == '.' {
			if start == i {
				return
			}
			start = i + 1
		}
		i++
	}

	if start == i {
		return
	}

	return v[:i], v[i:], true
}

// isIdentChar reports whether c is a valid identifier character
// in prerelease/build metadata ([0-9A-Za-z-]).
func isIdentChar(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-'
}

// isBadNum reports whether v is a numeric identifier
// with leading zeroes (which are invalid).
func isBadNum(v string) bool {
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

	return x[:i], x[i:]
}
