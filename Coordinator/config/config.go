package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ZoneID       string           `yaml:"zone_id"`
	Server       ServerConfig     `yaml:"server"`
	Gossip       GossipConfig     `yaml:"gossip"`
	Management   ManagementConfig `yaml:"management"`
	Health       HealthConfig     `yaml:"health"`
	Distribution DistConfig       `yaml:"distribution"`
}

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr"`
}

type GossipConfig struct {
	BindPort int      `yaml:"bind_port"`
	Seeds    []string `yaml:"seeds"`
	NodeName string   `yaml:"node_name"`
}

type ManagementConfig struct {
	URL          string `yaml:"url"`
	AuthToken    string `yaml:"auth_token"`
	SyncInterval string `yaml:"sync_interval"`
}

type HealthConfig struct {
	ProbeInterval    string `yaml:"probe_interval"`
	HeartbeatTimeout string `yaml:"heartbeat_timeout"`
	ProbeTimeout     string `yaml:"probe_timeout"`
	FailThreshold    int    `yaml:"fail_threshold"`
}

type DistConfig struct {
	HashPartitions    int    `yaml:"hash_partitions"`
	RebalanceDebounce string `yaml:"rebalance_debounce"`
}

func (c *Config) SyncInterval() time.Duration {
	d, _ := time.ParseDuration(c.Management.SyncInterval)
	if d == 0 {
		return 30 * time.Second
	}
	return d
}

func (c *Config) ProbeInterval() time.Duration {
	d, _ := time.ParseDuration(c.Health.ProbeInterval)
	if d == 0 {
		return 15 * time.Second
	}
	return d
}

func (c *Config) HeartbeatTimeout() time.Duration {
	d, _ := time.ParseDuration(c.Health.HeartbeatTimeout)
	if d == 0 {
		return 90 * time.Second
	}
	return d
}

func (c *Config) ProbeTimeout() time.Duration {
	d, _ := time.ParseDuration(c.Health.ProbeTimeout)
	if d == 0 {
		return 5 * time.Second
	}
	return d
}

func (c *Config) RebalanceDebounce() time.Duration {
	d, _ := time.ParseDuration(c.Distribution.RebalanceDebounce)
	if d == 0 {
		return 5 * time.Second
	}
	return d
}

func (c *Config) HashPartitions() int {
	if c.Distribution.HashPartitions <= 0 {
		return 160
	}
	return c.Distribution.HashPartitions
}

func (c *Config) Validate() error {
	if c.ZoneID == "" {
		return fmt.Errorf("zone_id 不能为空")
	}
	if c.Management.URL != "" {
		u, err := url.Parse(c.Management.URL)
		if err != nil {
			return fmt.Errorf("management.url 格式错误: %w", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("management.url 必须为 http 或 https")
		}
		host := u.Hostname()
		if isPrivateOrLoopback(host) {
			// 管理平台的地址允许为内网地址（运维配置场景），但记录警告
			_ = host // 内网地址在运维场景中是预期的
		}
	}
	return nil
}

func isPrivateOrLoopback(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		// 尝试解析为 host:port
		h, _, err := net.SplitHostPort(host)
		if err != nil {
			return strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")
		}
		ip = net.ParseIP(h)
		if ip == nil {
			return false
		}
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Gossip.NodeName == "" {
		cfg.Gossip.NodeName = "coordinator-" + cfg.ZoneID
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}
	return &cfg, nil
}
