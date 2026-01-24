package diff

import (
	"bufio"
	"strings"
)

// Hunk represents a change hunk in a diff.
type Hunk struct {
	StartLine int
	Lines     []Line
}

// LineType represents the type of a line in a hunk.
type LineType int

const (
	Context LineType = iota
	Addition
	Deletion
)

// Line represents a single line in a hunk.
type Line struct {
	Type    LineType
	Content string
}

// Diff represents a parsed diff for a file.
type Diff struct {
	File  string
	Hunks []Hunk
}

// Parse parses a V4A diff string.
// Supports:
// *** Update File: <path> ***
// *** Add File: <path> *** (treated as update with creates)
// @@ ... @@ headers
// + - space prefixes
func Parse(input string) ([]Diff, error) {
	var diffs []Diff
	var currentDiff *Diff
	var currentHunk *Hunk

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()

		// Check for file header
		if strings.HasPrefix(line, "*** Update File: ") || strings.HasPrefix(line, "*** Add File: ") {
			if currentDiff != nil {
				if currentHunk != nil {
					currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
					currentHunk = nil
				}
				diffs = append(diffs, *currentDiff)
			}
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				path := strings.TrimSuffix(parts[1], " ***")
				currentDiff = &Diff{File: path}
			}
			continue
		}

		if strings.HasPrefix(line, "*** End Patch") || strings.HasPrefix(line, "*** End of File") {
			if currentDiff != nil {
				if currentHunk != nil {
					currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
					currentHunk = nil
				}
				diffs = append(diffs, *currentDiff)
				currentDiff = nil
			}
			continue
		}

		if currentDiff == nil {
			continue // Skip preamble
		}

		// Check for hunk header
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
			}
			// Parse start line if needed, simple parser ignores complex validation for now
			currentHunk = &Hunk{}
			continue
		}

		// Parse content lines
		if currentHunk != nil {
			if strings.HasPrefix(line, "+") {
				currentHunk.Lines = append(currentHunk.Lines, Line{Type: Addition, Content: line[1:]})
			} else if strings.HasPrefix(line, "-") {
				currentHunk.Lines = append(currentHunk.Lines, Line{Type: Deletion, Content: line[1:]})
			} else if strings.HasPrefix(line, " ") {
				currentHunk.Lines = append(currentHunk.Lines, Line{Type: Context, Content: line[1:]})
			} else if line == "" {
				// Empty line often treated as context or ignore?
				// Context usually has space. Empty string might be empty context line.
				currentHunk.Lines = append(currentHunk.Lines, Line{Type: Context, Content: ""})
			}
		}
	}

	if currentDiff != nil {
		if currentHunk != nil {
			currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
		}
		diffs = append(diffs, *currentDiff)
	}

	return diffs, scanner.Err()
}
