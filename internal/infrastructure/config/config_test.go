package config

import (
	"testing"
)

func TestConfig_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	c := FromEnv()
	c.LastDeviceID = "index:1"
	c.AudioFeedback = true
	c.OutputDeviceID = "index:2"
	c.BeepVolume = 0.5
	if err := SaveFile(path, c); err != nil { t.Fatalf("save: %v", err) }
	// load file; ensure values persisted regardless of env defaults
	got, err := LoadFile(path)
	if err != nil { t.Fatalf("load: %v", err) }
	if got.LastDeviceID != c.LastDeviceID || got.OutputDeviceID != c.OutputDeviceID || got.BeepVolume != c.BeepVolume || !got.AudioFeedback {
		t.Fatalf("mismatch after load: %+v", got)
	}
	if got.Language == "" { t.Fatalf("expected Language to be set") }
}
