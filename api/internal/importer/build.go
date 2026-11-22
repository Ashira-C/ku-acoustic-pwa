package importer

import (
	"fmt"
	"slices"
	"strings"
)

const (
	RoleMember  = "MEMBER"
	RoleManager = "MANAGER"
	RoleAdmin   = "ADMIN"
)

// Member is a cleaned, validated user ready to be written to the database.
type Member struct {
	StudentID string
	Name      string
	Nickname  string
	Parts     []string
	Role      string
}

// Issue is a problem tied to a line in one of the input files.
type Issue struct {
	Source  string // "members.csv" or "corrections.csv"
	Line    int
	Message string
}

func (i Issue) String() string {
	return fmt.Sprintf("%s:%d: %s", i.Source, i.Line, i.Message)
}

// Result is the outcome of Build. Members must not be imported while Errors is non-empty.
type Result struct {
	Members  []Member
	Errors   []Issue
	Warnings []Issue
}

var allowedFields = map[string]bool{
	"student_id": true, "name": true, "nickname": true, "parts": true, "role": true,
}

var allowedRoles = map[string]bool{RoleMember: true, RoleManager: true, RoleAdmin: true}

// Build cleans raw rows, applies corrections and validates the result.
// It is a pure function (no files, no database), which keeps it easy to unit test.
func Build(rows []RawRow, corrections []Correction) Result {
	var res Result
	errorf := func(src string, line int, format string, args ...any) {
		res.Errors = append(res.Errors, Issue{src, line, fmt.Sprintf(format, args...)})
	}
	warnf := func(src string, line int, format string, args ...any) {
		res.Warnings = append(res.Warnings, Issue{src, line, fmt.Sprintf(format, args...)})
	}

	// 1. Index corrections by the student ID written in members.csv.
	fixes := make(map[string]map[string]Correction)
	for _, c := range corrections {
		switch {
		case !allowedFields[c.Field]:
			errorf("corrections.csv", c.Line, "unknown field %q", c.Field)
			continue
		case c.Value == "":
			errorf("corrections.csv", c.Line, "%s for %s has no value (unresolved TODO?)", c.Field, c.MatchStudentID)
			continue
		}
		if fixes[c.MatchStudentID] == nil {
			fixes[c.MatchStudentID] = make(map[string]Correction)
		}
		if _, dup := fixes[c.MatchStudentID][c.Field]; dup {
			errorf("corrections.csv", c.Line, "second %s correction for %s", c.Field, c.MatchStudentID)
			continue
		}
		fixes[c.MatchStudentID][c.Field] = c
	}
	matched := make(map[string]bool)

	// 2. Clean each row and apply its corrections.
	bySID := make(map[string]int) // final student ID -> index in res.Members
	for _, row := range rows {
		key := strings.TrimSpace(row.StudentID)
		fix := fixes[key]
		if fix != nil {
			matched[key] = true
		}
		pick := func(field, original string) string {
			if c, ok := fix[field]; ok {
				return c.Value
			}
			return original
		}

		m := Member{
			StudentID: pick("student_id", key),
			Name:      cleanText(pick("name", row.Name)),
			Nickname:  cleanText(pick("nickname", row.Nickname)),
			Role:      pick("role", RoleMember),
		}

		parts, err := parseParts(pick("parts", row.Parts))
		if err != nil {
			errorf("members.csv", row.Line, "%s: %v", m.StudentID, err)
			continue
		}
		m.Parts = parts

		// 3. Validate.
		ok := true
		if !studentIDPattern.MatchString(m.StudentID) {
			errorf("members.csv", row.Line, "invalid student ID %q", m.StudentID)
			ok = false
		}
		if m.Name == "" {
			errorf("members.csv", row.Line, "%s: name is empty", m.StudentID)
			ok = false
		}
		if m.Nickname == "" {
			errorf("members.csv", row.Line, "%s: nickname is empty", m.StudentID)
			ok = false
		}
		if !allowedRoles[m.Role] {
			errorf("members.csv", row.Line, "%s: invalid role %q", m.StudentID, m.Role)
			ok = false
		}
		if !ok {
			continue
		}
		if len(m.Parts) == 0 {
			warnf("members.csv", row.Line, "%s (%s): no part", m.StudentID, m.Nickname)
		}
		if !strings.Contains(m.Name, " ") {
			warnf("members.csv", row.Line, "%s (%s): name has no surname?", m.StudentID, m.Nickname)
		}

		// 4. Duplicates: identical rows are skipped, conflicting ones are errors.
		if i, seen := bySID[m.StudentID]; seen {
			if sameMember(res.Members[i], m) {
				warnf("members.csv", row.Line, "%s: exact duplicate row skipped", m.StudentID)
			} else {
				errorf("members.csv", row.Line, "%s: appears twice with different data", m.StudentID)
			}
			continue
		}
		bySID[m.StudentID] = len(res.Members)
		res.Members = append(res.Members, m)
	}

	// 5. A correction that matches no row is almost always a typo in its student ID.
	for _, c := range corrections {
		if allowedFields[c.Field] && c.Value != "" && !matched[c.MatchStudentID] {
			errorf("corrections.csv", c.Line, "student ID %s not found in members.csv", c.MatchStudentID)
		}
	}

	return res
}

func sameMember(a, b Member) bool {
	return a.StudentID == b.StudentID && a.Name == b.Name && a.Nickname == b.Nickname &&
		a.Role == b.Role && slices.Equal(a.Parts, b.Parts)
}