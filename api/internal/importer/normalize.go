package importer

import (
	"fmt"
	"regexp"
	"strings"
)

// partAliases maps every spelling found in the roster (lower-cased)
// to the canonical part name seeded in the parts table.
var partAliases = map[string]string{
	"vocal":                         "Vocal",
	"guitar":                        "Guitar",
	"bass":                          "Bass",
	"keyboard":                      "Keyboard",
	"percussion":                    "Percussion",
	"extra":                         "Extra",
	"technician":                    "Technician",
	"technician & sound engineer":   "Technician",
	"sound engineer and technician": "Technician",
}

var (
	studentIDPattern = regexp.MustCompile(`^\d{10}$`)
	partSeparators   = regexp.MustCompile(`[/|,]`)
)

// cleanText trims, collapses repeated spaces and fixes a common Thai typo:
// two "เ" typed in a row instead of "แ" (they look identical on screen).
func cleanText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.ReplaceAll(s, "เเ", "แ")
}

// parseParts turns "Guitar / Keyboard" into ["Guitar", "Keyboard"].
// An empty string means "no part" (reported as a warning, not an error).
func parseParts(raw string) ([]string, error) {
	var parts []string
	seen := make(map[string]bool)

	for _, p := range partSeparators.Split(raw, -1) {
		key := strings.ToLower(cleanText(p))
		if key == "" {
			continue
		}
		canonical, ok := partAliases[key]
		if !ok {
			return nil, fmt.Errorf("unknown part %q", strings.TrimSpace(p))
		}
		if !seen[canonical] {
			seen[canonical] = true
			parts = append(parts, canonical)
		}
	}
	return parts, nil
}