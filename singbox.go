package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// GenerateSingboxConfig creates a sing-box configuration that proxies
// specified processes and/or domains. All other traffic goes direct.
func GenerateSingboxConfig(cfg *Config) map[string]interface{} {
	logInfo("generating sing-box config...")

	// Build outbound based on config mode
	var proxyOutbound map[string]interface{}
	var excludeIPs []string

	if cfg.HasUpstream() {
		// Upstream proxy mode (e.g. local V2RayX/Clash)
		logInfo("mode: upstream proxy (%s %s:%d)", cfg.Upstream.Type, cfg.Upstream.Host, cfg.Upstream.Port)
		proxyOutbound = map[string]interface{}{
			"type":        "socks",
			"tag":         "proxy",
			"server":      cfg.Upstream.Host,
			"server_port": cfg.Upstream.Port,
		}
		if cfg.Upstream.Type == "http" {
			proxyOutbound["type"] = "http"
		}
		// Exclude upstream proxy from TUN to prevent loop
		if cfg.Upstream.Host != "127.0.0.1" && cfg.Upstream.Host != "localhost" {
			excludeIPs = append(excludeIPs, cfg.Upstream.Host+"/32")
		}
	} else if len(cfg.Nodes) > 0 {
		// Direct node mode
		node := cfg.Nodes[cfg.SelectedNode]
		logInfo("mode: direct node (%s %s:%d, type=%s)", node.Name, node.Server, node.Port, node.Type)

		// Resolve server IP to exclude from TUN
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

		proxyOutbound = map[string]interface{}{
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
	} else {
		logError("no upstream proxy or nodes configured!")
		return nil
	}

	// Target processes
	targetProcesses := cfg.GetTargetProcesses()
	logInfo("target processes (%d):", len(targetProcesses))
	for _, p := range targetProcesses {
		logInfo("  - %s", p)
	}

	// Target domains
	targetDomains := cfg.GetTargetDomains()
	logInfo("target domains (%d):", len(targetDomains))
	for _, d := range targetDomains {
		logInfo("  - %s", d)
	}

	// Route rules:
	// 1. Private IPs -> direct
	// 2. Target domains -> proxy (regardless of process)
	// 3. Target processes -> proxy
	// 4. Everything else -> direct
	routeRules := []map[string]interface{}{
		{"ip_is_private": true, "outbound": "direct"},
	}

	if len(targetDomains) > 0 {
		routeRules = append(routeRules, map[string]interface{}{
			"domain_suffix": targetDomains,
			"outbound":      "proxy",
		})
	}

	if len(targetProcesses) > 0 {
		routeRules = append(routeRules, map[string]interface{}{
			"process_name": targetProcesses,
			"outbound":     "proxy",
		})
	}

	// DNS rules:
	// - Target domains use remote DNS (via proxy)
	// - Target processes use remote DNS (via proxy)
	// - Everything else uses system DNS
	systemDNS := detectSystemDNS()
	logInfo("system DNS: %s", systemDNS)

	dnsRules := []map[string]interface{}{
		{"outbound": "any", "server": "dns-direct"},
	}
	if len(targetDomains) > 0 {
		dnsRules = append(dnsRules, map[string]interface{}{
			"domain_suffix": targetDomains,
			"server":        "dns-remote",
		})
	}
	if len(targetProcesses) > 0 {
		dnsRules = append(dnsRules, map[string]interface{}{
			"process_name": targetProcesses,
			"server":       "dns-remote",
		})
	}

	// Outbounds
	outbounds := []map[string]interface{}{
		proxyOutbound,
		{"type": "direct", "tag": "direct"},
	}

	// TUN inbound
	var excludeAddrs []string
	excludeAddrs = append(excludeAddrs, excludeIPs...)
	if systemDNS != "" && systemDNS != "127.0.0.1" {
		excludeAddrs = append(excludeAddrs, systemDNS+"/32")
		logInfo("excluding system DNS from TUN: %s", systemDNS)
	}
	logInfo("route_exclude_address: %v", excludeAddrs)

	tunInbound := map[string]interface{}{
		"type":                        "tun",
		"tag":                         "tun-in",
		"address":                     []string{"172.19.0.1/28"},
		"auto_route":                  true,
		"strict_route":                true,
		"stack":                       "gvisor",
		"sniff":                       true,
		"sniff_override_destination":  true,
		"route_exclude_address":       excludeAddrs,
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

	mode := "upstream"
	if !cfg.HasUpstream() {
		mode = "direct-node"
	}
	logInfo("sing-box config generated: mode=%s, route.final=direct, %d route rules, %d domain rules",
		mode, len(routeRules), len(targetDomains))
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

func formatUpstreamAddr(cfg *Config) string {
	if cfg.HasUpstream() {
		return fmt.Sprintf("%s://%s:%d", cfg.Upstream.Type, cfg.Upstream.Host, cfg.Upstream.Port)
	}
	return ""
}
