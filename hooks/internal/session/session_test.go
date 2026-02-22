package session

import (
	"testing"
)

func TestParseFilename_Valid(t *testing.T) {
	tests := []struct {
		filename string
		date     string
		shortID  string
	}{
		{"2026-02-01-session.tmp", "2026-02-01", "no-id"},
		{"2026-02-01-a1b2c3d4-session.tmp", "2026-02-01", "a1b2c3d4"},
		{"2024-12-25-abcdef12-session.tmp", "2024-12-25", "abcdef12"},
	}

	for _, tt := range tests {
		info := ParseFilename(tt.filename)
		if info == nil {
			t.Errorf("ParseFilename(%q) = nil, want non-nil", tt.filename)
			continue
		}
		if info.Date != tt.date {
			t.Errorf("ParseFilename(%q).Date = %q, want %q", tt.filename, info.Date, tt.date)
		}
		if info.ShortID != tt.shortID {
			t.Errorf("ParseFilename(%q).ShortID = %q, want %q", tt.filename, info.ShortID, tt.shortID)
		}
	}
}

func TestParseFilename_Invalid(t *testing.T) {
	tests := []string{
		"not-a-session.tmp",
		"2026-13-01-session.tmp",  // invalid month
		"2026-02-31-session.tmp",  // Feb 31
		"session.tmp",
		"2026-02-01.tmp",
		"",
	}

	for _, tt := range tests {
		if info := ParseFilename(tt); info != nil {
			t.Errorf("ParseFilename(%q) = %+v, want nil", tt, info)
		}
	}
}

func TestParseMetadata(t *testing.T) {
	content := `# Session: 2026-02-01
**Date:** 2026-02-01
**Started:** 09:00
**Last Updated:** 17:30

---

## Current State

### Completed
- [x] Task one
- [x] Task two

### In Progress
- [ ] Task three

### Notes for Next Session
Remember to review PR

### Context to Load
` + "```\n" + `src/main.ts
` + "```\n"

	md := ParseMetadata(content)

	if md.Title != "Session: 2026-02-01" {
		t.Errorf("Title = %q", md.Title)
	}
	if md.Date != "2026-02-01" {
		t.Errorf("Date = %q", md.Date)
	}
	if md.Started != "09:00" {
		t.Errorf("Started = %q", md.Started)
	}
	if md.LastUpdated != "17:30" {
		t.Errorf("LastUpdated = %q", md.LastUpdated)
	}
	if len(md.Completed) != 2 {
		t.Errorf("Completed count = %d, want 2", len(md.Completed))
	}
	if len(md.InProgress) != 1 {
		t.Errorf("InProgress count = %d, want 1", len(md.InProgress))
	}
	if md.Notes != "Remember to review PR" {
		t.Errorf("Notes = %q", md.Notes)
	}
	if md.Context != "src/main.ts" {
		t.Errorf("Context = %q", md.Context)
	}
}

func TestParseMetadata_Empty(t *testing.T) {
	md := ParseMetadata("")
	if md.Title != "" || md.Date != "" {
		t.Error("ParseMetadata('') should return empty metadata")
	}
}

func TestGetStats(t *testing.T) {
	content := "line1\nline2\n### Completed\n- [x] Done\n\n### In Progress\n- [ ] WIP\n"
	stats := GetStats(content)

	if stats.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", stats.TotalItems)
	}
	if stats.CompletedItems != 1 {
		t.Errorf("CompletedItems = %d, want 1", stats.CompletedItems)
	}
	if stats.LineCount != 8 {
		t.Errorf("LineCount = %d, want 8", stats.LineCount)
	}
}

func TestFormatSize_Missing(t *testing.T) {
	s := FormatSize("/nonexistent/path")
	if s != "0 B" {
		t.Errorf("FormatSize(missing) = %q, want '0 B'", s)
	}
}

