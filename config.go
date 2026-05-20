package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// NodeConfig defines a proxy node (vmess/shadowsocks)
type NodeConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // "vmess" or "shadowsocks"
	Server   string `json:"server"`   // server address
	Port     int    `json:"port"`     // server port
	UUID     string `json:"uuid,omitempty"`
	Security string `json:"security,omitempty"`
	Method   string `json:"method,omitempty"`
	Password string `json:"password,omitempty"`
}

// Config is the main configuration
type Config struct {
	Nodes        []NodeConfig `json:"nodes"`
	SelectedNode int          `json:"selected_node"`
	SubscribeURL string       `json:"subscribe_url,omitempty"`

	// TargetProcesses lists process names to proxy.
	// Default: Antigravity-related processes.
	TargetProcesses []string `json:"target_processes,omitempty"`

	// LogLevel for sing-box: trace/debug/info/warn/error
	LogLevel string `json:"log_level,omitempty"`
}

// DefaultTargetProcesses are the Antigravity-related process names on macOS
var DefaultTargetProcesses = []string{
	"Antigravity",
	"Antigravity Helper",
	"Antigravity Helper (Renderer)",
	"Antigravity Helper (GPU)",
	"language_server",
	"Electron",
	"Antigravity IDE Helper",
	"Antigravity IDE Helper (Plugin)",
	"Antigravity IDE Helper (Renderer)",
	"Antigravity IDE Helper (GPU)",
	"language_server_macos_arm",
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".antigravity-proxy")
	os.MkdirAll(dir, 0755)
	return dir
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func LoadConfig() (*Config, error) {
	c := &Config{
		LogLevel: "info",
	}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		// No config yet
		return c, nil
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	return c, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0644)
}

func (c *Config) GetTargetProcesses() []string {
	if len(c.TargetProcesses) > 0 {
		return c.TargetProcesses
	}
	return DefaultTargetProcesses
}
