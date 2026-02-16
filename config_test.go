package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	c := loadConfig()
	if !c.MergeBoxBorders {
		t.Error("MergeBoxBorders should default to true")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config", "pixl")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config"), []byte("merge-box-borders = false\n"), 0644)

	orig := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	defer os.Setenv("HOME", orig)

	c := loadConfig()
	if c.MergeBoxBorders {
		t.Error("MergeBoxBorders should be false when config says false")
	}
}

func TestLoadConfigIgnoresComments(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config", "pixl")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config"), []byte("# a comment\nmerge-box-borders = true\n"), 0644)

	t.Setenv("HOME", dir)

	c := loadConfig()
	if !c.MergeBoxBorders {
		t.Error("MergeBoxBorders should be true")
	}
}
