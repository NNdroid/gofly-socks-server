package config

import (
	"errors"
	"gofly/pkg/engine"
	"time"
)

type IPluginConfig interface {
	Check() error
}

type Server struct {
}

type Config struct {
	VTunSettings      VTunConfig      `yaml:"vTunSettings"`
	Tun2SocksSettings engine.Key      `yaml:"socksSettings"`
	WebSocketSettings WebSocketConfig `yaml:"wsSettings"`
	RealitySettings   RealityConfig   `yaml:"realitySettings"`
}

type VTunConfig struct {
	LocalAddr  string `yaml:"local_addr"`
	Key        string `yaml:"key"`
	Protocol   string `yaml:"protocol"`
	Obfs       bool   `yaml:"obfs"`
	Compress   bool   `yaml:"compress"`
	MTU        int    `yaml:"mtu"`
	Timeout    int    `yaml:"timeout"` //Unit second
	BufferSize int    `yaml:"buffer_size"`
	Verbose    bool   `yaml:"verbose"`
}

type WebSocketConfig struct {
	Path                      string `yaml:"path"`
	TLSCertificateFilePath    string `yaml:"tls_certificate_file_path"`
	TLSCertificateKeyFilePath string `yaml:"tls_certificate_key_file_path"`
}

func (c *WebSocketConfig) Check() error {
	if c.Path == "" {
		c.Path = "/"
	}
	return nil
}

type RealityConfig struct {
	ShortID     []string `yaml:"short_id"`
	ServerNames []string `yaml:"server_names"`
	Dest        string   `yaml:"dest"`
	PrivateKey  string   `yaml:"private_key"`
	Debug       bool     `yaml:"debug"`
}

func (c *RealityConfig) Check() error {
	if len(c.ShortID) == 0 {
		return errors.New("shortId can not empty")
	}
	if len(c.ServerNames) == 0 {
		return errors.New("serverNames can not empty")
	}
	if c.Dest == "" {
		return errors.New("dest can not empty")
	}
	if c.PrivateKey == "" {
		return errors.New("privateKey can not empty")
	}
	return nil
}

func (config *Config) setDefault() {
	if config.VTunSettings.BufferSize == 0 {
		config.VTunSettings.BufferSize = 65535
	}
	if config.VTunSettings.Protocol == "" {
		config.VTunSettings.Protocol = "ws"
	}
	if config.VTunSettings.MTU == 0 {
		config.VTunSettings.MTU = 1500
	}
	if config.VTunSettings.Timeout == 0 {
		config.VTunSettings.Timeout = 60
	}

	if config.Tun2SocksSettings.MTU == 0 {
		config.Tun2SocksSettings.MTU = 1500
	}
	if config.Tun2SocksSettings.UDPTimeout == 0 {
		config.Tun2SocksSettings.UDPTimeout = 60 * time.Second
	}
	if config.Tun2SocksSettings.Device == "" {
		config.Tun2SocksSettings.Device = "tun0"
	}
	if !config.RealitySettings.Debug {
		config.RealitySettings.Debug = config.VTunSettings.Verbose
	}
}
