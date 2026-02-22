// Package session provides session file CRUD, parsing, and metadata.
package session

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oguzsh/claudey/internal/fileutil"
	"github.com/oguzsh/claudey/internal/hookio"
	"github.com/oguzsh/claudey/internal/platform"
)

// Session filename pattern: YYYY-MM-DD-[short-id]-session.tmp
var filenameRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?:-([a-z0-9]{8,}))?-session\.tmp$`)

// FileInfo holds parsed session filename metadata.
type FileInfo struct {
	Filename string
	ShortID  string
	Date     string
	Datetime time.Time
}

// ParseFilename extracts metadata from a session filename.
func ParseFilename(filename string) *FileInfo {
	m := filenameRegex.FindStringSubmatch(filename)
	if m == nil {
		return nil
	}

	dateStr := m[1]
	parts := strings.Split(dateStr, "-")
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])

	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
	}

	// Validate date is real (not Feb 31, etc.)
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	if int(d.Month()) != month || d.Day() != day {
		return nil
	}

	shortID := m[2]
	if shortID == "" {
		shortID = "no-id"
	}

	return &FileInfo{
		Filename: filename,
		ShortID:  shortID,
		Date:     dateStr,
		Datetime: d,
	}
}

// Path returns the full path to a session file.
func Path(filename string) string {
	return filepath.Join(platform.SessionsDir(), filename)
}

// Content reads and returns session file content.
func Content(sessionPath string) string {
	c, ok := fileutil.ReadFile(sessionPath)
	if !ok {
		return ""
	}
	return c
}

// Metadata holds parsed session content metadata.
type Metadata struct {
	Title       string
	Date        string
	Started     string
	LastUpdated string
	Completed   []string
	InProgress  []string
	Notes       string
	Context     string
}

var (
	titleRe     = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	dateRe      = regexp.MustCompile(`\*\*Date:\*\*\s*(\d{4}-\d{2}-\d{2})`)
	startedRe   = regexp.MustCompile(`\*\*Started:\*\*\s*([\d:]+)`)
	updatedRe   = regexp.MustCompile(`\*\*Last Updated:\*\*\s*([\d:]+)`)
	completedRe = regexp.MustCompile(`### Completed\s*\n([\s\S]*?)(?:###|\n\n|$)`)
	progressRe  = regexp.MustCompile(`### In Progress\s*\n([\s\S]*?)(?:###|\n\n|$)`)
	notesRe     = regexp.MustCompile(`### Notes for Next Session\s*\n([\s\S]*?)(?:###|\n\n|$)`)
	contextRe   = regexp.MustCompile("### Context to Load\\s*\n```\n([\\s\\S]*?)```")
	checkboxXRe = regexp.MustCompile(`- \[x\]\s*(.+)`)
	checkboxRe  = regexp.MustCompile(`- \[ \]\s*(.+)`)
)

// ParseMetadata parses session metadata from markdown content.
func ParseMetadata(content string) Metadata {
	md := Metadata{}
	if content == "" {
		return md
	}

	if m := titleRe.FindStringSubmatch(content); m != nil {
		md.Title = strings.TrimSpace(m[1])
	}
	if m := dateRe.FindStringSubmatch(content); m != nil {
		md.Date = m[1]
	}
	if m := startedRe.FindStringSubmatch(content); m != nil {
		md.Started = m[1]
	}
	if m := updatedRe.FindStringSubmatch(content); m != nil {
		md.LastUpdated = m[1]
	}

	if m := completedRe.FindStringSubmatch(content); m != nil {
		for _, item := range checkboxXRe.FindAllStringSubmatch(m[1], -1) {
			md.Completed = append(md.Completed, strings.TrimSpace(item[1]))
		}
	}

	if m := progressRe.FindStringSubmatch(content); m != nil {
		for _, item := range checkboxRe.FindAllStringSubmatch(m[1], -1) {
			md.InProgress = append(md.InProgress, strings.TrimSpace(item[1]))
		}
	}

	if m := notesRe.FindStringSubmatch(content); m != nil {
		md.Notes = strings.TrimSpace(m[1])
	}

	if m := contextRe.FindStringSubmatch(content); m != nil {
		md.Context = strings.TrimSpace(m[1])
	}

	return md
}

// Stats holds session statistics.
type Stats struct {
	TotalItems    int
	CompletedItems int
	InProgressItems int
	LineCount     int
	HasNotes      bool
	HasContext     bool
}

// GetStats calculates statistics for session content.
func GetStats(content string) Stats {
	md := ParseMetadata(content)
	lineCount := 0
	if content != "" {
		lineCount = len(strings.Split(content, "\n"))
	}
	return Stats{
		TotalItems:      len(md.Completed) + len(md.InProgress),
		CompletedItems:  len(md.Completed),
		InProgressItems: len(md.InProgress),
		LineCount:       lineCount,
		HasNotes:        md.Notes != "",
		HasContext:       md.Context != "",
	}
}

// SessionEntry holds a full session entry with file stats.
type SessionEntry struct {
	FileInfo
	SessionPath  string
	HasContent   bool
	Size         int64
	ModifiedTime time.Time
	CreatedTime  time.Time
}

// ListResult holds a paginated list of sessions.
type ListResult struct {
	Sessions []SessionEntry
	Total    int
	Offset   int
	Limit    int
	HasMore  bool
}

// ListAll returns all sessions with optional filtering and pagination.
func ListAll(limit, offset int, dateFilter, search string) ListResult {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	sessionsDir := platform.SessionsDir()
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return ListResult{Limit: limit, Offset: offset}
	}

	var sessions []SessionEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}

		info := ParseFilename(entry.Name())
		if info == nil {
			continue
		}

		if dateFilter != "" && info.Date != dateFilter {
			continue
		}
		if search != "" && !strings.Contains(info.ShortID, search) {
			continue
		}

		sessionPath := filepath.Join(sessionsDir, entry.Name())
		stat, err := os.Stat(sessionPath)
		if err != nil {
			continue
		}

		sessions = append(sessions, SessionEntry{
			FileInfo:     *info,
			SessionPath:  sessionPath,
			HasContent:   stat.Size() > 0,
			Size:         stat.Size(),
			ModifiedTime: stat.ModTime(),
			CreatedTime:  stat.ModTime(), // Go doesn't reliably expose birth time
		})
	}

	// Sort newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModifiedTime.After(sessions[j].ModifiedTime)
	})

	total := len(sessions)
	end := offset + limit
	if end > total {
		end = total
	}
	start := offset
	if start > total {
		start = total
	}

	return ListResult{
		Sessions: sessions[start:end],
		Total:    total,
		Offset:   offset,
		Limit:    limit,
		HasMore:  end < total,
	}
}

// FindByID finds a session by short ID or filename.
func FindByID(sessionID string, includeContent bool) *SessionEntry {
	sessionsDir := platform.SessionsDir()
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}

		info := ParseFilename(entry.Name())
		if info == nil {
			continue
		}

		shortIDMatch := len(sessionID) > 0 && info.ShortID != "no-id" && strings.HasPrefix(info.ShortID, sessionID)
		filenameMatch := entry.Name() == sessionID || entry.Name() == sessionID+".tmp"
		noIDMatch := info.ShortID == "no-id" && entry.Name() == sessionID+"-session.tmp"

		if !shortIDMatch && !filenameMatch && !noIDMatch {
			continue
		}

		sessionPath := filepath.Join(sessionsDir, entry.Name())
		stat, err := os.Stat(sessionPath)
		if err != nil {
			return nil
		}

		se := &SessionEntry{
			FileInfo:     *info,
			SessionPath:  sessionPath,
			Size:         stat.Size(),
			ModifiedTime: stat.ModTime(),
			CreatedTime:  stat.ModTime(),
		}

		return se
	}

	return nil
}

// Title returns the title from a session file.
func Title(sessionPath string) string {
	content := Content(sessionPath)
	md := ParseMetadata(content)
	if md.Title != "" {
		return md.Title
	}
	return "Untitled Session"
}

// FormatSize formats file size in human-readable format.
func FormatSize(sessionPath string) string {
	stat, err := os.Stat(sessionPath)
	if err != nil {
		return "0 B"
	}
	size := stat.Size()
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

// WriteContent writes session content to a file.
func WriteContent(sessionPath, content string) bool {
	if err := fileutil.WriteFile(sessionPath, content); err != nil {
		hookio.Logf("[SessionManager] Error writing session: %s", err)
		return false
	}
	return true
}

// AppendContent appends content to a session file.
func AppendContent(sessionPath, content string) bool {
	if err := fileutil.AppendFile(sessionPath, content); err != nil {
		hookio.Logf("[SessionManager] Error appending to session: %s", err)
		return false
	}
	return true
}

// Delete removes a session file.
func Delete(sessionPath string) bool {
	if !fileutil.Exists(sessionPath) {
		return false
	}
	if err := os.Remove(sessionPath); err != nil {
		hookio.Logf("[SessionManager] Error deleting session: %s", err)
		return false
	}
	return true
}

// Exists checks if a session file exists.
func Exists(sessionPath string) bool {
	info, err := os.Stat(sessionPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}



