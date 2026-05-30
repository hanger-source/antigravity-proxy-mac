package main

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DetectedProxy holds the result of system proxy detection
type DetectedProxy struct {
	Type   string // "socks5" or "http"
	Host   string
	Port   int
	Source string // where it was detected from
}

// detectSystemProxy tries to detect a working local proxy by:
// 1. Checking macOS system proxy settings (scutil --proxy)
// 2. Probing common proxy ports
func detectSystemProxy() *DetectedProxy {
	// Step 1: Check system proxy via scutil
	if p := detectFromScutil(); p != nil {
		if probePort(p.Host, p.Port) {
			logInfo("detected system proxy: %s %s:%d (source: %s)", p.Type, p.Host, p.Port, p.Source)
			return p
		}
		logInfo("system proxy %s:%d configured but not responding", p.Host, p.Port)
	}

	// Step 2: Probe common local proxy ports
	commonPorts := []struct {
		port    int
		typ     string
		comment string
	}{
		{7890, "socks5", "Clash"},
		{7891, "socks5", "Clash SOCKS"},
		{1080, "socks5", "common SOCKS5"},
		{1087, "socks5", "common SOCKS5 alt"},
		{13658, "socks5", "V2RayX"},
		{1081, "http", "common HTTP"},
		{8080, "http", "common HTTP alt"},
		{8118, "http", "Privoxy"},
	}

	for _, p := range commonPorts {
		if probePort("127.0.0.1", p.port) {
			logInfo("detected proxy on port %d (%s) via port scan", p.port, p.comment)
			return &DetectedProxy{
				Type:   p.typ,
				Host:   "127.0.0.1",
				Port:   p.port,
				Source: fmt.Sprintf("port-scan: %s", p.comment),
			}
		}
	}

	return nil
}

// detectFromScutil parses macOS scutil --proxy output
func detectFromScutil() *DetectedProxy {
	out, err := exec.Command("scutil", "--proxy").Output()
	if err != nil {
		return nil
	}
	lines := string(out)

	// Check SOCKS proxy
	if matchEnabled(lines, "SOCKSEnable") {
		host := matchValue(lines, "SOCKSProxy")
		port := matchIntValue(lines, "SOCKSPort")
		if host != "" && port > 0 {
			return &DetectedProxy{Type: "socks5", Host: host, Port: port, Source: "scutil SOCKS"}
		}
	}

	// Check HTTPS proxy
	if matchEnabled(lines, "HTTPSEnable") {
		host := matchValue(lines, "HTTPSProxy")
		port := matchIntValue(lines, "HTTPSPort")
		if host != "" && port > 0 {
			return &DetectedProxy{Type: "http", Host: host, Port: port, Source: "scutil HTTPS"}
		}
	}

	// Check HTTP proxy
	if matchEnabled(lines, "HTTPEnable") {
		host := matchValue(lines, "HTTPProxy")
		port := matchIntValue(lines, "HTTPPort")
		if host != "" && port > 0 {
			return &DetectedProxy{Type: "http", Host: host, Port: port, Source: "scutil HTTP"}
		}
	}

	return nil
}

var reKV = regexp.MustCompile(`(\w+)\s*:\s*(.+)`)

func matchEnabled(text, key string) bool {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if m := reKV.FindStringSubmatch(line); m != nil {
			if m[1] == key && strings.TrimSpace(m[2]) == "1" {
				return true
			}
		}
	}
	return false
}

func matchValue(text, key string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if m := reKV.FindStringSubmatch(line); m != nil {
			if m[1] == key {
				return strings.TrimSpace(m[2])
			}
		}
	}
	return ""
}

func matchIntValue(text, key string) int {
	v := matchValue(text, key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

// probePort checks if a TCP port is accepting connections
func probePort(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
