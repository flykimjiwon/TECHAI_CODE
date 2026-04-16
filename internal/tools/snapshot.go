package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kimjiwon/tgc/internal/config"
)

// SnapshotEntry records a single file backup.
type SnapshotEntry struct {
	OriginalPath string
	SnapshotPath string
	Timestamp    time.Time
}

const maxSnapshots = 50

// snapshotDir returns ~/.tgc/snapshots/ (auto-created).
func snapshotDir() string {
	dir := filepath.Join(config.ConfigDir(), "snapshots")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// CreateSnapshot backs up a file before modification.
// Returns the snapshot path or empty string if file doesn't exist yet.
func CreateSnapshot(absPath string) string {
	data, err := os.ReadFile(absPath)
	if err != nil {
		// File doesn't exist yet (new file) — nothing to snapshot
		return ""
	}

	ts := time.Now().Format("20060102-150405.000")
	baseName := filepath.Base(absPath)
	snapName := fmt.Sprintf("%s_%s", ts, baseName)
	snapPath := filepath.Join(snapshotDir(), snapName)

	if err := os.WriteFile(snapPath, data, 0644); err != nil {
		config.DebugLog("[SNAPSHOT] failed to create snapshot: %v", err)
		return ""
	}

	// Write metadata (original path) alongside snapshot
	metaPath := snapPath + ".meta"
	_ = os.WriteFile(metaPath, []byte(absPath), 0644)

	config.DebugLog("[SNAPSHOT] created %s → %s (%d bytes)", absPath, snapPath, len(data))

	// Prune old snapshots
	pruneSnapshots()

	return snapPath
}

// SnapshotAndWrite creates a snapshot of the existing file, then writes new content.
func SnapshotAndWrite(absPath string, content []byte) error {
	CreateSnapshot(absPath)
	return os.WriteFile(absPath, content, 0644)
}

// ListSnapshots returns recent snapshots, newest first.
func ListSnapshots(limit int) []SnapshotEntry {
	dir := snapshotDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var snapshots []SnapshotEntry
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".meta") {
			continue
		}
		snapPath := filepath.Join(dir, e.Name())
		metaPath := snapPath + ".meta"
		origPath := ""
		if data, err := os.ReadFile(metaPath); err == nil {
			origPath = string(data)
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		snapshots = append(snapshots, SnapshotEntry{
			OriginalPath: origPath,
			SnapshotPath: snapPath,
			Timestamp:    info.ModTime(),
		})
	}

	// Sort newest first
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	if limit > 0 && len(snapshots) > limit {
		snapshots = snapshots[:limit]
	}
	return snapshots
}

// UndoLast restores the N most recent snapshots. Returns a summary.
func UndoLast(count int) string {
	if count <= 0 {
		count = 1
	}
	snapshots := ListSnapshots(count)
	if len(snapshots) == 0 {
		return "되돌릴 스냅샷이 없습니다."
	}

	var results []string
	for _, snap := range snapshots {
		if snap.OriginalPath == "" {
			results = append(results, fmt.Sprintf("⚠ %s: 원본 경로 없음 (스킵)", filepath.Base(snap.SnapshotPath)))
			continue
		}
		data, err := os.ReadFile(snap.SnapshotPath)
		if err != nil {
			results = append(results, fmt.Sprintf("⚠ %s: 읽기 실패", filepath.Base(snap.SnapshotPath)))
			continue
		}
		if err := os.WriteFile(snap.OriginalPath, data, 0644); err != nil {
			results = append(results, fmt.Sprintf("⚠ %s → %s: 복원 실패", filepath.Base(snap.SnapshotPath), snap.OriginalPath))
			continue
		}
		// Remove used snapshot + meta
		_ = os.Remove(snap.SnapshotPath)
		_ = os.Remove(snap.SnapshotPath + ".meta")
		results = append(results, fmt.Sprintf("✓ %s 복원됨", snap.OriginalPath))
		config.DebugLog("[UNDO] restored %s from %s", snap.OriginalPath, snap.SnapshotPath)
	}
	return strings.Join(results, "\n")
}

// FormatSnapshotList formats snapshots for display.
func FormatSnapshotList(limit int) string {
	snapshots := ListSnapshots(limit)
	if len(snapshots) == 0 {
		return "스냅샷이 없습니다."
	}
	var lines []string
	for i, s := range snapshots {
		ts := s.Timestamp.Format("01/02 15:04:05")
		orig := s.OriginalPath
		if orig == "" {
			orig = "(알 수 없음)"
		}
		lines = append(lines, fmt.Sprintf("  %d. [%s] %s", i+1, ts, orig))
	}
	return strings.Join(lines, "\n")
}

// pruneSnapshots removes oldest snapshots when over maxSnapshots.
func pruneSnapshots() {
	all := ListSnapshots(0)
	if len(all) <= maxSnapshots {
		return
	}
	// Remove oldest (list is newest-first)
	for _, old := range all[maxSnapshots:] {
		_ = os.Remove(old.SnapshotPath)
		_ = os.Remove(old.SnapshotPath + ".meta")
	}
	config.DebugLog("[SNAPSHOT] pruned %d old snapshots", len(all)-maxSnapshots)
}
