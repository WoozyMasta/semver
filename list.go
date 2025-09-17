package semver

import "sort"

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
		ai = ls[i].Canonical()
	}

	aj := ls[j].Original
	if aj == "" {
		aj = ls[j].Canonical()
	}

	return ai < aj
}

// Sort sorts the list in ascending semver order.
func (ls List) Sort() {
	sort.Sort(ls)
}
