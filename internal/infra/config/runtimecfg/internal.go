package runtimecfg

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type InternalConfig struct {
	Gateway  InternalGatewayConfig  `toml:"gateway" mapstructure:"gateway" json:"gateway" yaml:"gateway"`
	Dispatch InternalDispatchConfig `toml:"dispatch" mapstructure:"dispatch" json:"dispatch" yaml:"dispatch"`
	Agent    InternalAgentConfig    `toml:"agent" mapstructure:"agent" json:"agent" yaml:"agent"`
}

type InternalGatewayConfig struct {
	Addr           string `toml:"addr" mapstructure:"addr" json:"addr" yaml:"addr"`
	AuthorizedKeys string `toml:"authorized_keys" mapstructure:"authorized_keys" json:"authorized_keys" yaml:"authorized_keys"`
	SSH            bool   `toml:"ssh" mapstructure:"ssh" json:"ssh" yaml:"ssh"`
	HostKey        string `toml:"host_key" mapstructure:"host_key" json:"host_key" yaml:"host_key"`
}

type InternalDispatchConfig struct {
	FetchTimeout    uint `toml:"fetch_timeout" mapstructure:"fetch_timeout" json:"fetch_timeout" yaml:"fetch_timeout"`
	DownloadTimeout uint `toml:"download_timeout" mapstructure:"download_timeout" json:"download_timeout" yaml:"download_timeout"`
}

type InternalAgentConfig struct {
	Gateway     string `toml:"gateway" mapstructure:"gateway" json:"gateway" yaml:"gateway"`
	Name        string `toml:"name" mapstructure:"name" json:"name" yaml:"name"`
	Concurrency uint   `toml:"concurrency" mapstructure:"concurrency" json:"concurrency" yaml:"concurrency"`
	PrivateKey  string `toml:"private_key" mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	SSH         bool   `toml:"ssh" mapstructure:"ssh" json:"ssh" yaml:"ssh"`
	KnownHosts  string `toml:"known_hosts" mapstructure:"known_hosts" json:"known_hosts" yaml:"known_hosts"`
}

const maxDurationSeconds = uint64((1<<63 - 1) / int64(time.Second))

func validateSeconds(value uint, name string) error {
	if value == 0 || uint64(value) > maxDurationSeconds {
		return fmt.Errorf("%s must be positive and fit a duration in seconds", name)
	}
	return nil
}

func (c InternalGatewayConfig) Validate() error {
	if c.Addr != "" {
		if _, _, err := net.SplitHostPort(c.Addr); err != nil {
			return fmt.Errorf("internal.gateway.addr must be a host:port listen address")
		}
		if strings.TrimSpace(c.AuthorizedKeys) == "" {
			return fmt.Errorf("internal.gateway.authorized_keys is required for the internal gateway")
		}
	}
	return nil
}

func (c InternalDispatchConfig) Validate() error {
	if err := validateSeconds(c.FetchTimeout, "internal.dispatch.fetch_timeout"); err != nil {
		return err
	}
	return validateSeconds(c.DownloadTimeout, "internal.dispatch.download_timeout")
}

func (c InternalAgentConfig) Validate() error {
	if c.Gateway == "" {
		return nil
	}
	host, port, err := net.SplitHostPort(c.Gateway)
	if err != nil || strings.TrimSpace(host) == "" {
		return fmt.Errorf("internal.agent.gateway must be a host:port address")
	}
	if parsed, err := strconv.ParseUint(port, 10, 16); err != nil || parsed == 0 {
		return fmt.Errorf("internal.agent.gateway has an invalid port")
	}
	if strings.TrimSpace(c.PrivateKey) == "" {
		return fmt.Errorf("internal.agent.private_key is required for an internal agent")
	}
	return nil
}
