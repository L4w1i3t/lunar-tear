package memory

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"lunar-tear/server/internal/store"
)

// 0 = unlimited (no pruning); set to a positive value to cap stored snapshots.
const maxSnapshots = 0

func snapshotPath(dir string, sceneId int32) string {
	return filepath.Join(dir, fmt.Sprintf("scene_%d.json", sceneId))
}

func saveSnapshot(user *store.UserState, dir string) {
	sceneId := user.MainQuest.CurrentQuestSceneId
	if sceneId == 0 {
		return
	}
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		log.Printf("[snapshot] marshal error for scene=%d: %v", sceneId, err)
		return
	}
	path := snapshotPath(dir, sceneId)
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("[snapshot] write error for scene=%d: %v", sceneId, err)
		return
	}
	log.Printf("[snapshot] saved scene=%d (%d bytes)", sceneId, len(data))
	pruneSnapshots(dir)
}

// pruneSnapshots deletes the oldest scene_*.json files in dir, keeping only the maxSnapshots most recently modified ones.
// When maxSnapshots is 0, pruning is disabled and all snapshots are kept.
func pruneSnapshots(dir string) {
	if maxSnapshots == 0 {
		return
	}
	matches, err := filepath.Glob(filepath.Join(dir, "scene_*.json"))
	if err != nil || len(matches) <= maxSnapshots {
		return
	}

	type entry struct {
		path string
		mod  int64
	}
	entries := make([]entry, 0, len(matches))
	for _, p := range matches {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		entries = append(entries, entry{path: p, mod: info.ModTime().UnixNano()})
	}

	// Sort newest first.
	sort.Slice(entries, func(i, j int) bool { return entries[i].mod > entries[j].mod })

	for _, e := range entries[maxSnapshots:] {
		if err := os.Remove(e.path); err != nil {
			log.Printf("[snapshot] prune error %q: %v", e.path, err)
			continue
		}
		log.Printf("[snapshot] pruned old snapshot %q", filepath.Base(e.path))
	}
}

// LatestSnapshotSceneId scans dir for scene_*.json files and returns the scene ID
// of the most recently modified snapshot. Returns (0, false) if none are found.
func LatestSnapshotSceneId(dir string) (int32, bool) {
	matches, err := filepath.Glob(filepath.Join(dir, "scene_*.json"))
	if err != nil || len(matches) == 0 {
		return 0, false
	}
	var latestPath string
	var latestMod int64
	for _, p := range matches {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if mt := info.ModTime().UnixNano(); mt > latestMod {
			latestMod = mt
			latestPath = p
		}
	}
	if latestPath == "" {
		return 0, false
	}
	base := filepath.Base(latestPath)
	name := strings.TrimSuffix(strings.TrimPrefix(base, "scene_"), ".json")
	id, err := strconv.ParseInt(name, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(id), true
}

func loadSnapshot(dir string, sceneId int32) (*store.UserState, error) {
	path := snapshotPath(dir, sceneId)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot scene=%d: %w", sceneId, err)
	}
	var user store.UserState
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot scene=%d: %w", sceneId, err)
	}
	user.EnsureMaps()
	return &user, nil
}
