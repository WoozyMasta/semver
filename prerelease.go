package semver

// comparePrerelease compares two prerelease strings (without the leading '-').
// Rules:
//   - Empty string (no prerelease) has higher precedence than any prerelease.
//   - Identifiers are compared dot by dot.
//   - Numeric identifiers compare numerically; non-numeric compare lexicographically (ASCII).
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

// isNum reports whether v consists entirely of digits.
func isNum(v string) bool {
	i := 0
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}

	return i == len(v)
}
