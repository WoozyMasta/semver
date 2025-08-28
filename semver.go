package semver

import (
	"sort"
	"strconv"
	"strings"
)

// Semver represents a parsed semantic version string
type Semver struct {
	// Original the raw input string (may be without "v")
	Original string

	// Canonical normalized form "vMAJOR.MINOR.PATCH[-PRERELEASE]" (build metadata stripped)
	Canonical string

	// Major numeric component (normalized, no leading zeros)
	Major int

	// Minor numeric component (normalized, no leading zeros)
	Minor int

	// Patch numeric component (normalized, no leading zeros)
	Patch int

	// Prerelease optional pre-release part (without leading '-')
	Prerelease string

	// Build optional build metadata (without leading '+')
	Build string

	// Valid indicates successful parsing
	Valid bool
}

// Parse parses a version string into a Semver struct.
// The input may include or omit a leading "v".
// Returns (Semver, true) if valid, otherwise (Semver{Valid:false}, false).
func Parse(s string) (Semver, bool) {
	orig := s
	p, ok := parseInternal(s)
	if !ok {
		return Semver{Original: orig, Valid: false}, false
	}

	maj, _ := strconv.Atoi(p.Major)
	min, _ := strconv.Atoi(p.Minor)
	pat, _ := strconv.Atoi(p.Patch)

	canon := "v" + p.Major + "." + p.Minor + "." + p.Patch
	if p.Prerelease != "" {
		canon += p.Prerelease // already starts with '-'
	}

	return Semver{
		Original:   orig,
		Canonical:  canon,
		Major:      maj,
		Minor:      min,
		Patch:      pat,
		Prerelease: strings.TrimPrefix(p.Prerelease, "-"),
		Build:      strings.TrimPrefix(p.Build, "+"),
		Valid:      true,
	}, true
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
	switch {
	case v.Prerelease == "" && w.Prerelease == "":
		return 0
	case v.Prerelease == "" && w.Prerelease != "":
		return 1
	case v.Prerelease != "" && w.Prerelease == "":
		return -1
	default:
		return comparePrerelease(v.Prerelease, w.Prerelease)
	}
}

// Canon returns the canonical string form "vMAJOR.MINOR.PATCH[-PRERELEASE]",
// with build metadata stripped. Empty if invalid.
func (v Semver) Canon() string { return v.Canonical }

// MajorStr returns the major-only string, e.g. "v2".
// Empty if invalid.
func (v Semver) MajorStr() string {
	if !v.Valid {
		return ""
	}

	return "v" + strconv.Itoa(v.Major)
}

// MajorMinorStr returns the major.minor string, e.g. "v2.1".
// Empty if invalid.
func (v Semver) MajorMinorStr() string {
	if !v.Valid {
		return ""
	}

	return "v" + strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor)
}

// Pre returns the prerelease identifier (without leading '-'), or "".
func (v Semver) Pre() string {
	return v.Prerelease
}

// BuildMeta returns the build metadata (without leading '+'), or "".
func (v Semver) BuildMeta() string {
	return v.Build
}

// String implements fmt.Stringer. It returns the canonical form.
func (v Semver) String() string {
	return v.Canonical
}

// List is a slice of Semver values that implements sort.Interface.
// Elements are ordered by semantic version precedence with a
// lexicographic tie-breaker on Original.
type List []Semver

// Len implements sort.Interface.
func (ls List) Len() int {
	return len(ls)
}

// Swap implements sort.Interface.
func (ls List) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

// Less implements sort.Interface.
// It orders by semantic version precedence; if two values compare equal
// it falls back to lexicographic order of Original (or Canon if empty).
func (ls List) Less(i, j int) bool {
	c := ls[i].Compare(ls[j])
	if c != 0 {
		return c < 0
	}

	ai := ls[i].Original
	if ai == "" {
		ai = ls[i].Canon()
	}

	aj := ls[j].Original
	if aj == "" {
		aj = ls[j].Canon()
	}

	return ai < aj
}

// Sort sorts the list in ascending semver order.
func (ls List) Sort() {
	sort.Sort(ls)
}

// Max returns the greater of two Semver values.
// Invalid values are always smaller than valid ones.
func (v Semver) Max(w Semver) Semver {
	if v.Compare(w) >= 0 {
		return v
	}

	return w
}
