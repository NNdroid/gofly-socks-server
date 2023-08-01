package engine

import "time"

type Key struct {
	MTU                      int           `yaml:"mtu"`
	Proxy                    string        `yaml:"proxy"`
	Device                   string        `yaml:"device"`
	TCPModerateReceiveBuffer bool          `yaml:"tcp-moderate-receive-buffer"`
	TCPSendBufferSize        string        `yaml:"tcp-send-buffer-size"`
	TCPReceiveBufferSize     string        `yaml:"tcp-receive-buffer-size"`
	UDPTimeout               time.Duration `yaml:"udp-timeout"`
}
