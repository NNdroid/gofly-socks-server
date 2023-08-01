package engine

import (
	"errors"
	"github.com/docker/go-units"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"os/exec"
	"strings"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
	"github.com/xjasonlyu/tun2socks/v2/engine/mirror"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
)

var (
	_engineMu sync.Mutex

	// _defaultKey holds the default key for the engine.
	_defaultKey *Key

	// _defaultProxy holds the default proxy for the engine.
	_defaultProxy proxy.Proxy

	// _defaultDevice holds the default device for the engine.
	_defaultDevice device.Device

	// _defaultStack holds the default stack for the engine.
	_defaultStack *stack.Stack
)

// Start starts the default engine up.
func Start() {
	if err := start(); err != nil {
		log.Fatalf("[ENGINE] failed to start: %v", err)
	}
}

// Stop shuts the default engine down.
func Stop() {
	if err := stop(); err != nil {
		log.Fatalf("[ENGINE] failed to stop: %v", err)
	}
}

// Insert loads *Key to the default engine.
func Insert(k *Key) {
	_engineMu.Lock()
	_defaultKey = k
	_engineMu.Unlock()
}

func start() error {
	_engineMu.Lock()
	if _defaultKey == nil {
		return errors.New("empty key")
	}

	for _, f := range []func(*Key) error{
		netStack,
	} {
		if err := f(_defaultKey); err != nil {
			return err
		}
	}
	_engineMu.Unlock()
	return nil
}

func stop() (err error) {
	_engineMu.Lock()
	if _defaultDevice != nil {
		err = _defaultDevice.Close()
	}
	if _defaultStack != nil {
		_defaultStack.Close()
		_defaultStack.Wait()
	}
	_engineMu.Unlock()
	return err
}

func execCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return errors.New("empty command")
	}
	_, err := exec.Command(parts[0], parts[1:]...).Output()
	return err
}

func SetLogLevel(debug bool) {
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.SilentLevel)
	}
}

func netStack(k *Key) (err error) {
	if k.Proxy == "" {
		return errors.New("empty proxy")
	}
	if k.Device == "" {
		return errors.New("empty device")
	}

	if _defaultProxy, err = parseProxy(k.Proxy); err != nil {
		return
	}
	proxy.SetDialer(_defaultProxy)

	if _defaultDevice, err = parseDevice(k.Device, uint32(k.MTU)); err != nil {
		return
	}

	var opts []option.Option
	if k.TCPModerateReceiveBuffer {
		opts = append(opts, option.WithTCPModerateReceiveBuffer(true))
	}

	if k.TCPSendBufferSize != "" {
		size, err := units.RAMInBytes(k.TCPSendBufferSize)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTCPSendBufferSize(int(size)))
	}

	if k.TCPReceiveBufferSize != "" {
		size, err := units.RAMInBytes(k.TCPReceiveBufferSize)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTCPReceiveBufferSize(int(size)))
	}

	if _defaultStack, err = core.CreateStack(&core.Config{
		LinkEndpoint:     _defaultDevice,
		TransportHandler: &mirror.Tunnel{},
		Options:          opts,
	}); err != nil {
		return
	}

	log.Infof(
		"[STACK] %s://%s <-> %s://%s",
		_defaultDevice.Type(), _defaultDevice.Name(),
		_defaultProxy.Proto(), _defaultProxy.Addr(),
	)
	return nil
}
