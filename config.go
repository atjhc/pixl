package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	MergeBoxBorders bool
}

func loadConfig() Config {
	c := Config{
		MergeBoxBorders: true,
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return c
	}

	f, err := os.Open(filepath.Join(home, ".config", "pixl", "config"))
	if err != nil {
		return c
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "merge-box-borders":
			c.MergeBoxBorders = val == "true"
		}
	}

	return c
}
