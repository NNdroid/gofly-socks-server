package tun

import (
	"fmt"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
	"gofly/pkg/device"
	"sync"
)

const Driver = "tun"
const offset = 0

func (t *TUN) Type() string {
	return Driver
}

var _ device.Device = (*TUN)(nil)

type TUN struct {
	*iobased.Endpoint

	mtu    uint32
	name   string
	offset int

	rMutex sync.Mutex
	wMutex sync.Mutex
}

var _rChan = make(chan []byte, 3000)
var _wChan = make(chan []byte, 3000)

func WriteToTun(packet []byte) (int, error) {
	n := len(packet)
	_wChan <- packet
	return n, nil
}

func ReadFromTun(packet []byte) (int, error) {
	b := <-_rChan
	n := len(b)
	copy(packet[:n], b)
	return n, nil
}

func Open(name string, mtu uint32) (_ device.Device, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("open tun: %v", r)
		}
	}()

	t := &TUN{
		name:   name,
		mtu:    mtu,
		offset: offset,
	}

	ep, err := iobased.New(t, t.mtu, offset)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	t.Endpoint = ep

	return t, nil
}

func (t *TUN) Read(packet []byte) (int, error) {
	t.rMutex.Lock()
	defer t.rMutex.Unlock()
	b := <-_wChan
	n := len(b)
	copy(packet[:n], b)
	return n, nil
}

func (t *TUN) Write(packet []byte) (int, error) {
	n := len(packet)
	_rChan <- packet
	return n, nil
}

func (t *TUN) Name() string {
	return t.name
}

func (t *TUN) Close() error {
	defer t.Endpoint.Close()
	return nil
}
