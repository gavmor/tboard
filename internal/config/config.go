package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConfigFile representation
type ConfigFile struct {
	Theme       string         `json:"theme"`
	DefaultPort int            `json:"default_port"`
	Targets     []ConfigTarget `json:"targets"`
}

type ConfigTarget struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// LoadConfigFile attempts to parse JSON or YAML config files.
func LoadConfigFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err == nil {
		return &cfg, nil
	}

	lines := strings.Split(string(data), "\n")
	cfg = ConfigFile{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "theme:") {
			cfg.Theme = strings.TrimSpace(strings.TrimPrefix(line, "theme:"))
		} else if strings.HasPrefix(line, "default_port:") {
			if p, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "default_port:"))); err == nil {
				cfg.DefaultPort = p
			}
		} else if strings.HasPrefix(line, "- host:") || strings.HasPrefix(line, "- ") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				host := strings.TrimSpace(parts[1])
				port := 80
				if len(parts) >= 3 {
					if p, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
						port = p
					}
				}
				if host != "" {
					cfg.Targets = append(cfg.Targets, ConfigTarget{Host: host, Port: port})
				}
			}
		}
	}

	return &cfg, nil
}

// DiscoverConfigFile searches default candidate locations for config files.
func DiscoverConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		"tboard.yaml",
		"tboard.json",
		filepath.Join(homeDir, ".config", "tboard", "config.yaml"),
		filepath.Join(homeDir, ".config", "tboard", "config.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
