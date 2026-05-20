package main

import (
	"bufio"
	"net"
	"os"
	"strings"
)

// GenerateSingboxConfig creates a sing-box configuration that only proxies
// Antigravity-related processes via direct node connection. All other traffic goes direct.
func GenerateSingboxConfig(cfg *Config) map[string]interface{} {
	logInfo("generating sing-box config...")

	if len(cfg.Nodes) == 0 {
		logError("no nodes configured!")
		return nil
	}

	node := cfg.Nodes[cfg.SelectedNode]
	logInfo("using node: %s (%s:%d, type=%s)", node.Name, node.Server, node.Port, node.Type)

	// Resolve server IP to exclude from TUN (prevent routing loop)
	var excludeIPs []string
	server := node.Server
	if net.ParseIP(server) == nil {
		if addrs, err := net.LookupHost(server); err == nil && len(addrs) > 0 {
			logInfo("resolved %s -> %v", server, addrs)
			for _, addr := range addrs {
				excludeIPs = append(excludeIPs, addr+"/32")
			}
			server = addrs[0]
		} else {
			logError("failed to resolve %s: %v", server, err)
		}
	} else {
		excludeIPs = append(excludeIPs, server+"/32")
	}

	// Build outbound for the node
	proxyOutbound := map[string]interface{}{
		"type":        node.Type,
		"tag":         "proxy",
		"server":      server,
		"server_port": node.Port,
	}
	if node.Type == "vmess" {
		proxyOutbound["uuid"] = node.UUID
		proxyOutbound["security"] = "auto"
		proxyOutbound["authenticated_length"] = true
		proxyOutbound["packet_encoding"] = "xudp"
	} else if node.Type == "shadowsocks" {
		proxyOutbound["method"] = node.Method
		proxyOutbound["password"] = node.Password
	}

	// Target processes
	targetProcesses := cfg.GetTargetProcesses()
	logInfo("target processes (%d):", len(targetProcesses))
	for _, p := range targetProcesses {
		logInfo("  - %s", p)
	}

	// Route rules:
	// 1. Private IPs -> direct
	// 2. Antigravity processes -> proxy
	// 3. Everything else -> direct
	routeRules := []map[string]interface{}{
		{"ip_is_private": true, "outbound": "direct"},
		{"process_name": targetProcesses, "outbound": "proxy"},
	}

	// DNS: Antigravity processes use remote DNS (via proxy), everything else uses system DNS
	systemDNS := detectSystemDNS()
	logInfo("system DNS: %s", systemDNS)

	dnsRules := []map[string]interface{}{
		{"outbound": "any", "server": "dns-direct"},
		{"process_name": targetProcesses, "server": "dns-remote"},
	}

	// Outbounds
	outbounds := []map[string]interface{}{
		proxyOutbound,
		{"type": "direct", "tag": "direct"},
	}

	// TUN inbound
	// Exclude: proxy server IPs + system DNS (prevent routing loop & DNS loop)
	var excludeAddrs []string
	excludeAddrs = append(excludeAddrs, excludeIPs...)
	if systemDNS != "" && systemDNS != "127.0.0.1" {
		excludeAddrs = append(excludeAddrs, systemDNS+"/32")
		logInfo("excluding system DNS from TUN: %s", systemDNS)
	}
	logInfo("route_exclude_address: %v", excludeAddrs)

	tunInbound := map[string]interface{}{
		"type":                       "tun",
		"tag":                        "tun-in",
		"address":                    []string{"172.19.0.1/28"},
		"auto_route":                true,
		"strict_route":              true,
		"stack":                     "gvisor",
		"sniff":                     true,
		"sniff_override_destination": false,
		"route_exclude_address":     excludeAddrs,
	}

	// Log level
	logLevel := cfg.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}

	result := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     logLevel,
			"timestamp": true,
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{"tag": "dns-remote", "address": "tcp://1.1.1.1", "detour": "proxy"},
				{"tag": "dns-direct", "address": systemDNS, "detour": "direct"},
			},
			"rules":             dnsRules,
			"final":             "dns-direct",
			"strategy":          "prefer_ipv4",
			"independent_cache": true,
		},
		"inbounds":  []map[string]interface{}{tunInbound},
		"outbounds": outbounds,
		"route": map[string]interface{}{
			"auto_detect_interface": true,
			"rules":                 routeRules,
			"final":                 "direct",
		},
	}

	logInfo("sing-box config generated: node=%s, route.final=direct, %d route rules",
		node.Name, len(routeRules))
	return result
}

func detectSystemDNS() string {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "223.5.5.5"
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && net.ParseIP(fields[1]) != nil && !strings.Contains(fields[1], ":") {
				return fields[1]
			}
		}
	}
	return "223.5.5.5"
}
