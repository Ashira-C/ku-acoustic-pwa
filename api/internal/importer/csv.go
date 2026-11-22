package importer

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Column headers in the club's roster sheet ("รายชื่อทั้งหมด").
const (
	colNickname  = "ชื่อเล่น"
	colName      = "ชื่อ - สกุล"
	colStudentID = "รหัสนิสิต"
	colParts     = "พาร์ท"
)

// RawRow is one roster line exactly as exported, before any cleaning.
type RawRow struct {
	Line      int // CSV line number, used in error messages
	StudentID string
	Name      string
	Nickname  string
	Parts     string
}

// Correction is one manual fix from corrections.csv.
type Correction struct {
	Line           int
	MatchStudentID string // student ID as written in members.csv (before fixes)
	Field          string // student_id | name | nickname | parts | role
	Value          string
}

// ReadMembers parses the roster CSV. Columns are found by header name,
// so their order in the sheet does not matter; extra columns are ignored.
func ReadMembers(r io.Reader) ([]RawRow, error) {
	records, err := readAll(r)
	if err != nil {
		return nil, err
	}

	col, err := headerIndex(records[0], colNickname, colName, colStudentID, colParts)
	if err != nil {
		return nil, err
	}

	rows := make([]RawRow, 0, len(records)-1)
	for i, rec := range records[1:] {
		if isBlank(rec) {
			continue
		}
		rows = append(rows, RawRow{
			Line:      i + 2, // +1 for the header, +1 for 1-based lines
			StudentID: field(rec, col[colStudentID]),
			Name:      field(rec, col[colName]),
			Nickname:  field(rec, col[colNickname]),
			Parts:     field(rec, col[colParts]),
		})
	}
	return rows, nil
}

// ReadCorrections parses corrections.csv (match_student_id,field,value,note).
func ReadCorrections(r io.Reader) ([]Correction, error) {
	records, err := readAll(r)
	if err != nil {
		return nil, err
	}

	col, err := headerIndex(records[0], "match_student_id", "field", "value")
	if err != nil {
		return nil, err
	}

	out := make([]Correction, 0, len(records)-1)
	for i, rec := range records[1:] {
		if isBlank(rec) {
			continue
		}
		out = append(out, Correction{
			Line:           i + 2,
			MatchStudentID: strings.TrimSpace(field(rec, col["match_student_id"])),
			Field:          strings.TrimSpace(field(rec, col["field"])),
			Value:          strings.TrimSpace(field(rec, col["value"])),
		})
	}
	return out, nil
}

func readAll(r io.Reader) ([][]string, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // rows may have different lengths

	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("csv is empty")
	}

	// Excel's "CSV UTF-8" adds a byte-order mark before the first header.
	records[0][0] = strings.TrimPrefix(records[0][0], "\ufeff")
	return records, nil
}

func headerIndex(header []string, required ...string) (map[string]int, error) {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	for _, name := range required {
		if _, ok := idx[name]; !ok {
			return nil, fmt.Errorf("missing column %q in header", name)
		}
	}
	return idx, nil
}

func field(rec []string, i int) string {
	if i < len(rec) {
		return rec[i]
	}
	return ""
}

func isBlank(rec []string) bool {
	for _, v := range rec {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}