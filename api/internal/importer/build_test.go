package importer

import (
	"slices"
	"strings"
	"testing"
)

// All data here is fake. Real member data never goes into the repo.

func TestParseParts(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{name: "single", raw: "Vocal", want: []string{"Vocal"}},
		{name: "case and spaces", raw: "  technician ", want: []string{"Technician"}},
		{name: "long technician alias", raw: "Sound Engineer and Technician", want: []string{"Technician"}},
		{name: "two parts with slash", raw: "Guitar / Keyboard", want: []string{"Guitar", "Keyboard"}},
		{name: "pipe and duplicate", raw: "Vocal|vocal|Bass", want: []string{"Vocal", "Bass"}},
		{name: "empty means no part", raw: "", want: nil},
		{name: "unknown part", raw: "Violin", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParts(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanText(t *testing.T) {
	if got := cleanText("  สมชาย   ใจดี "); got != "สมชาย ใจดี" {
		t.Errorf("spaces: got %q", got)
	}
	if got := cleanText("เเก้ว"); got != "แก้ว" {
		t.Errorf("double sara e: got %q", got)
	}
}

func TestBuild(t *testing.T) {
	rows := []RawRow{
		{Line: 2, StudentID: "6700000001", Name: "สมชาย ใจดี", Nickname: "ชาย1", Parts: "Vocal"},
		{Line: 3, StudentID: "6700000002", Name: "สมหญิง รักดี", Nickname: "หญิง", Parts: "Guitar / Keyboard"},
		{Line: 4, StudentID: "6700000002", Name: "สมหญิง รักดี", Nickname: "หญิง", Parts: "Guitar / Keyboard"}, // exact duplicate
		{Line: 5, StudentID: "broken-id", Name: "สมศักดิ์ มั่นคง", Nickname: "ศักดิ์", Parts: ""},
	}
	corrections := []Correction{
		{Line: 2, MatchStudentID: "6700000001", Field: "nickname", Value: "ชาย"},
		{Line: 3, MatchStudentID: "6700000001", Field: "role", Value: RoleManager},
		{Line: 4, MatchStudentID: "broken-id", Field: "student_id", Value: "6700000003"},
		{Line: 5, MatchStudentID: "broken-id", Field: "parts", Value: "Bass"},
	}

	res := Build(rows, corrections)

	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(res.Members) != 3 {
		t.Fatalf("got %d members, want 3", len(res.Members))
	}

	first := res.Members[0]
	if first.Nickname != "ชาย" || first.Role != RoleManager {
		t.Errorf("corrections not applied: %+v", first)
	}
	fixed := res.Members[2]
	if fixed.StudentID != "6700000003" || !slices.Equal(fixed.Parts, []string{"Bass"}) {
		t.Errorf("student_id/parts correction not applied: %+v", fixed)
	}
	if !hasIssue(res.Warnings, "exact duplicate") {
		t.Errorf("expected a duplicate warning, got %v", res.Warnings)
	}
}

func TestBuildErrors(t *testing.T) {
	tests := []struct {
		name        string
		rows        []RawRow
		corrections []Correction
		wantErr     string
	}{
		{
			name:    "invalid student ID",
			rows:    []RawRow{{Line: 2, StudentID: "12345", Name: "ก ข", Nickname: "ก", Parts: "Vocal"}},
			wantErr: "invalid student ID",
		},
		{
			name: "conflicting duplicate",
			rows: []RawRow{
				{Line: 2, StudentID: "6700000001", Name: "ก ข", Nickname: "ก", Parts: "Vocal"},
				{Line: 3, StudentID: "6700000001", Name: "ก ข", Nickname: "ก", Parts: "Bass"},
			},
			wantErr: "different data",
		},
		{
			name:        "correction for unknown student",
			rows:        []RawRow{{Line: 2, StudentID: "6700000001", Name: "ก ข", Nickname: "ก", Parts: "Vocal"}},
			corrections: []Correction{{Line: 2, MatchStudentID: "6799999999", Field: "nickname", Value: "x"}},
			wantErr:     "not found",
		},
		{
			name:        "unresolved TODO",
			rows:        []RawRow{{Line: 2, StudentID: "6700000001", Name: "ก ข", Nickname: "ก", Parts: "Vocal"}},
			corrections: []Correction{{Line: 2, MatchStudentID: "6700000001", Field: "nickname", Value: ""}},
			wantErr:     "no value",
		},
		{
			name:    "unknown part",
			rows:    []RawRow{{Line: 2, StudentID: "6700000001", Name: "ก ข", Nickname: "ก", Parts: "Violin"}},
			wantErr: "unknown part",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Build(tt.rows, tt.corrections)
			if !hasIssue(res.Errors, tt.wantErr) {
				t.Errorf("want error containing %q, got %v", tt.wantErr, res.Errors)
			}
		})
	}
}

func TestReadMembersByHeaderName(t *testing.T) {
	// BOM + columns in a different order + an extra column.
	csv := "\ufeffรหัสนิสิต,โทรศัพท์,ชื่อเล่น,ชื่อ - สกุล,พาร์ท\n6700000001,0800000000,ก,ก ข,Vocal\n,,,,\n"
	rows, err := ReadMembers(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].StudentID != "6700000001" || rows[0].Nickname != "ก" {
		t.Errorf("got %+v", rows)
	}
}

func hasIssue(issues []Issue, substr string) bool {
	for _, i := range issues {
		if strings.Contains(i.Message, substr) {
			return true
		}
	}
	return false
}