// Package aliases manages session alias storage in ~/.claude/session-aliases.json.
package aliases

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/oguzsh/claudey/internal/fileutil"
	"github.com/oguzsh/claudey/internal/hookio"
	"github.com/oguzsh/claudey/internal/platform"
)

const aliasVersion = "1.0"

var validAliasName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var reservedNames = map[string]bool{
	"list": true, "help": true, "remove": true,
	"delete": true, "create": true, "set": true,
}

// AliasEntry holds data for a single alias.
type AliasEntry struct {
	SessionPath string `json:"sessionPath"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	Title       string `json:"title,omitempty"`
}

// AliasFile is the on-disk JSON structure.
type AliasFile struct {
	Version  string                `json:"version"`
	Aliases  map[string]AliasEntry `json:"aliases"`
	Metadata struct {
		TotalCount  int    `json:"totalCount"`
		LastUpdated string `json:"lastUpdated"`
	} `json:"metadata"`
}

func aliasesPath() string {
	return filepath.Join(platform.ClaudeDir(), "session-aliases.json")
}

func defaultFile() AliasFile {
	f := AliasFile{
		Version: aliasVersion,
		Aliases: make(map[string]AliasEntry),
	}
	f.Metadata.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	return f
}

// Load reads aliases from disk.
func Load() AliasFile {
	path := aliasesPath()
	content, ok := fileutil.ReadFile(path)
	if !ok {
		return defaultFile()
	}

	var f AliasFile
	if err := json.Unmarshal([]byte(content), &f); err != nil {
		hookio.Logf("[Aliases] Error parsing aliases file: %s", err)
		return defaultFile()
	}

	if f.Aliases == nil {
		hookio.Log("[Aliases] Invalid aliases file structure, resetting")
		return defaultFile()
	}
	if f.Version == "" {
		f.Version = aliasVersion
	}

	return f
}

// Save writes aliases to disk with atomic write.
func Save(f AliasFile) bool {
	path := aliasesPath()
	tmpPath := path + ".tmp"
	bakPath := path + ".bak"

	f.Metadata.TotalCount = len(f.Aliases)
	f.Metadata.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		hookio.Logf("[Aliases] Error marshaling aliases: %s", err)
		return false
	}

	if err := platform.EnsureDir(filepath.Dir(path)); err != nil {
		hookio.Logf("[Aliases] Error creating directory: %s", err)
		return false
	}

	// Backup existing file
	if fileutil.Exists(path) {
		copyFile(path, bakPath)
	}

	// Write to temp file
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		hookio.Logf("[Aliases] Error writing temp file: %s", err)
		restoreBackup(bakPath, path)
		return false
	}

	// On Windows, remove destination before rename
	if runtime.GOOS == "windows" && fileutil.Exists(path) {
		os.Remove(path)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		hookio.Logf("[Aliases] Error renaming temp file: %s", err)
		restoreBackup(bakPath, path)
		cleanupTemp(tmpPath)
		return false
	}

	// Remove backup on success
	os.Remove(bakPath)
	return true
}

func copyFile(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	os.WriteFile(dst, data, 0o644)
}

func restoreBackup(bak, dst string) {
	if fileutil.Exists(bak) {
		data, err := os.ReadFile(bak)
		if err == nil {
			os.WriteFile(dst, data, 0o644)
			hookio.Log("[Aliases] Restored from backup")
		}
	}
}

func cleanupTemp(tmp string) {
	os.Remove(tmp)
}

// Resolve looks up an alias by name.
func Resolve(alias string) *AliasEntry {
	if alias == "" || !validAliasName.MatchString(alias) {
		return nil
	}

	f := Load()
	entry, ok := f.Aliases[alias]
	if !ok {
		return nil
	}
	return &entry
}

// SetResult holds the result of a Set operation.
type SetResult struct {
	Success     bool
	IsNew       bool
	Alias       string
	SessionPath string
	Title       string
	Error       string
}

// Set creates or updates an alias.
func Set(alias, sessionPath, title string) SetResult {
	if alias == "" {
		return SetResult{Error: "Alias name cannot be empty"}
	}
	if sessionPath == "" || strings.TrimSpace(sessionPath) == "" {
		return SetResult{Error: "Session path cannot be empty"}
	}
	if len(alias) > 128 {
		return SetResult{Error: "Alias name cannot exceed 128 characters"}
	}
	if !validAliasName.MatchString(alias) {
		return SetResult{Error: "Alias name must contain only letters, numbers, dashes, and underscores"}
	}
	if reservedNames[strings.ToLower(alias)] {
		return SetResult{Error: fmt.Sprintf("'%s' is a reserved alias name", alias)}
	}

	f := Load()
	existing, isExisting := f.Aliases[alias]
	now := time.Now().UTC().Format(time.RFC3339)

	entry := AliasEntry{
		SessionPath: sessionPath,
		UpdatedAt:   now,
	}
	if title != "" {
		entry.Title = title
	}
	if isExisting {
		entry.CreatedAt = existing.CreatedAt
	} else {
		entry.CreatedAt = now
	}

	f.Aliases[alias] = entry

	if !Save(f) {
		return SetResult{Error: "Failed to save alias"}
	}

	return SetResult{
		Success:     true,
		IsNew:       !isExisting,
		Alias:       alias,
		SessionPath: sessionPath,
		Title:       title,
	}
}

// AliasInfo holds info for listing.
type AliasInfo struct {
	Name        string
	SessionPath string
	CreatedAt   string
	UpdatedAt   string
	Title       string
}

// List returns all aliases, optionally filtered.
func List(search string, limit int) []AliasInfo {
	f := Load()

	var aliases []AliasInfo
	for name, entry := range f.Aliases {
		aliases = append(aliases, AliasInfo{
			Name:        name,
			SessionPath: entry.SessionPath,
			CreatedAt:   entry.CreatedAt,
			UpdatedAt:   entry.UpdatedAt,
			Title:       entry.Title,
		})
	}

	// Sort by updated time, newest first
	sort.Slice(aliases, func(i, j int) bool {
		ti := aliases[i].UpdatedAt
		if ti == "" {
			ti = aliases[i].CreatedAt
		}
		tj := aliases[j].UpdatedAt
		if tj == "" {
			tj = aliases[j].CreatedAt
		}
		return ti > tj
	})

	// Filter
	if search != "" {
		searchLower := strings.ToLower(search)
		var filtered []AliasInfo
		for _, a := range aliases {
			if strings.Contains(strings.ToLower(a.Name), searchLower) ||
				strings.Contains(strings.ToLower(a.Title), searchLower) {
				filtered = append(filtered, a)
			}
		}
		aliases = filtered
	}

	// Limit
	if limit > 0 && len(aliases) > limit {
		aliases = aliases[:limit]
	}

	return aliases
}

// DeleteResult holds the result of a delete operation.
type DeleteResult struct {
	Success            bool
	Alias              string
	DeletedSessionPath string
	Error              string
}

// Delete removes an alias.
func Delete(alias string) DeleteResult {
	f := Load()
	entry, ok := f.Aliases[alias]
	if !ok {
		return DeleteResult{Error: fmt.Sprintf("Alias '%s' not found", alias)}
	}

	delete(f.Aliases, alias)
	if !Save(f) {
		return DeleteResult{Error: "Failed to delete alias"}
	}

	return DeleteResult{
		Success:            true,
		Alias:              alias,
		DeletedSessionPath: entry.SessionPath,
	}
}

// Rename renames an alias.
func Rename(oldAlias, newAlias string) SetResult {
	if newAlias == "" {
		return SetResult{Error: "New alias name cannot be empty"}
	}
	if len(newAlias) > 128 {
		return SetResult{Error: "New alias name cannot exceed 128 characters"}
	}
	if !validAliasName.MatchString(newAlias) {
		return SetResult{Error: "New alias name must contain only letters, numbers, dashes, and underscores"}
	}
	if reservedNames[strings.ToLower(newAlias)] {
		return SetResult{Error: fmt.Sprintf("'%s' is a reserved alias name", newAlias)}
	}

	f := Load()
	entry, ok := f.Aliases[oldAlias]
	if !ok {
		return SetResult{Error: fmt.Sprintf("Alias '%s' not found", oldAlias)}
	}
	if _, exists := f.Aliases[newAlias]; exists {
		return SetResult{Error: fmt.Sprintf("Alias '%s' already exists", newAlias)}
	}

	delete(f.Aliases, oldAlias)
	entry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	f.Aliases[newAlias] = entry

	if !Save(f) {
		// Rollback
		f.Aliases[oldAlias] = entry
		delete(f.Aliases, newAlias)
		Save(f)
		return SetResult{Error: "Failed to save renamed alias — rolled back to original"}
	}

	return SetResult{
		Success:     true,
		Alias:       newAlias,
		SessionPath: entry.SessionPath,
	}
}

// ResolveSessionAlias resolves an alias to a session path, or returns the input as-is.
func ResolveSessionAlias(aliasOrID string) string {
	entry := Resolve(aliasOrID)
	if entry != nil {
		return entry.SessionPath
	}
	return aliasOrID
}

// ForSession returns all aliases pointing to a given session path.
func ForSession(sessionPath string) []AliasInfo {
	f := Load()
	var result []AliasInfo
	for name, entry := range f.Aliases {
		if entry.SessionPath == sessionPath {
			result = append(result, AliasInfo{
				Name:      name,
				CreatedAt: entry.CreatedAt,
				Title:     entry.Title,
			})
		}
	}
	return result
}

// Cleanup removes aliases pointing to non-existent sessions.
func Cleanup(sessionExists func(string) bool) (checked, removed int) {
	f := Load()
	var toRemove []string

	for name, entry := range f.Aliases {
		if !sessionExists(entry.SessionPath) {
			toRemove = append(toRemove, name)
		}
	}

	checked = len(f.Aliases)
	for _, name := range toRemove {
		delete(f.Aliases, name)
	}

	if len(toRemove) > 0 {
		Save(f)
	}

	return checked, len(toRemove)
}
