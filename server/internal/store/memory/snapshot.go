package memory

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lunar-tear/server/internal/store"
)

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
}

// parseSceneId extracts the numeric scene ID from a filename of the form "scene_<id>.json".
// Returns (0, false) if the name does not match the expected format.
func parseSceneId(name string) (int32, bool) {
	if !strings.HasPrefix(name, "scene_") || !strings.HasSuffix(name, ".json") {
		return 0, false
	}
	raw := strings.TrimSuffix(strings.TrimPrefix(name, "scene_"), ".json")
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(id), true
}

// LatestSnapshotSceneId scans dir for scene_*.json files and returns the scene ID
// of the most recently modified snapshot. Returns (0, false) if none are found.
func LatestSnapshotSceneId(dir string) (int32, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, false
	}
	var latestId int32
	var latestMod int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		id, ok := parseSceneId(e.Name())
		if !ok {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if mt := info.ModTime().UnixNano(); mt > latestMod {
			latestMod = mt
			latestId = id
		}
	}
	if latestId == 0 {
		return 0, false
	}
	return latestId, true
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
