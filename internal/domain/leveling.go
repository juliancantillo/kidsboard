package domain

import "math"

// XPForLevel returns the cumulative XP required to reach the given level.
// Level 1 starts at 0 XP. Curve: 50 * n * (n-1).
// Examples: L1=0, L2=100, L3=300, L4=600, L5=1000, L10=4500, L20=19000.
func XPForLevel(level int) int64 {
	if level <= 1 {
		return 0
	}
	n := int64(level)
	return 50 * n * (n - 1)
}

// LevelForXP returns the highest level whose XP threshold is <= xp.
func LevelForXP(xp int64) int {
	if xp <= 0 {
		return 1
	}
	// Inverse of XPForLevel: solve 50 n (n-1) <= xp.
	// n <= (1 + sqrt(1 + xp/12.5)) / 2
	n := math.Floor((1 + math.Sqrt(1+float64(xp)/12.5)) / 2)
	level := int(n)
	if level < 1 {
		return 1
	}
	return level
}

// XPForNextLevel returns the XP threshold of the level above the current.
func XPForNextLevel(currentXP int64) int64 {
	return XPForLevel(LevelForXP(currentXP) + 1)
}
