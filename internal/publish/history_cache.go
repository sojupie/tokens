package publish

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Bump when semantic indexing or comparison rules change.
const historyProcessingVersion = "semantic-history-1"

type historyCacheEntry struct {
	Key      string          `json:"key"`
	Checksum string          `json:"checksum"`
	Data     json.RawMessage `json:"data"`
}

func (config BuildConfig) historyCachePath(kind, identity string) string {
	directory := config.HistoryCacheDir
	if directory == "" {
		directory = filepath.Join(filepath.Dir(config.OutputDir), ".cache", "publication-history")
	}
	return filepath.Join(directory, sha256Hex([]byte(historyProcessingVersion+":"+kind+":"+identity))+".json")
}

func (config BuildConfig) readHistoryCache(kind, identity string, value any) bool {
	body, err := os.ReadFile(config.historyCachePath(kind, identity))
	if err != nil {
		return false
	}
	var entry historyCacheEntry
	if json.Unmarshal(body, &entry) != nil || entry.Key != historyProcessingVersion+":"+kind+":"+identity || entry.Checksum != sha256Hex(entry.Data) {
		return false
	}
	return json.Unmarshal(entry.Data, value) == nil
}

func (config BuildConfig) writeHistoryCache(kind, identity string, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	body, err := json.Marshal(historyCacheEntry{historyProcessingVersion + ":" + kind + ":" + identity, sha256Hex(data), data})
	if err != nil {
		return
	}
	path := config.historyCachePath(kind, identity)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		config.report("History cache unavailable: " + err.Error())
		return
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".history-*")
	if err != nil {
		config.report("History cache unavailable: " + err.Error())
		return
	}
	defer os.Remove(file.Name())
	_, err = file.Write(body)
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(file.Name(), path)
	}
	if err != nil {
		config.report("History cache write failed: " + err.Error())
	}
}
