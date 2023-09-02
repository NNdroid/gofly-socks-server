package basic

import (
	"context"
	"github.com/klauspost/compress/snappy"
	"github.com/patrickmn/go-cache"
	"gofly/pkg/cipher"
	"gofly/pkg/config"
	"gofly/pkg/logger"
	"gofly/pkg/statistics"
	"gofly/pkg/x/xcrypto"
	"gofly/pkg/x/xproto"
	"time"
)

type ServerForApi interface {
	Init()
	StartServerForApi()
}

func ContextOpened(_ctx context.Context) bool {
	select {
	case <-_ctx.Done():
		return false
	default:
		return true
	}
}

type Server struct {
	Config          *config.Config
	ReadFunc        func([]byte) (int, error)
	WriteFunc       func([]byte) int
	CTX             context.Context
	ConnectionCache *cache.Cache
	Statistics      *statistics.Statistics
	xp              *xcrypto.XCrypto
	authKey         *xproto.AuthKey
}

func (x *Server) Init() {
	cipher.SetKey(x.Config.VTunSettings.Key)
	x.authKey = xproto.ParseAuthKeyFromString(x.Config.VTunSettings.Key)
	x.xp = &xcrypto.XCrypto{}
	err := x.xp.Init(x.Config.VTunSettings.Key)
	if err != nil {
		logger.Logger.Sugar().Panicf("Init XCrypto failed: %s", err)
	}
}

func (x *Server) ConvertDstAddr(packet []byte) {
	//
}

func (x *Server) ConvertSrcAddr(packet []byte) {
	//
}

func (x *Server) BasicEncode(b []byte) ([]byte, error) {
	if x.Config.VTunSettings.Obfs {
		b = cipher.XOR(b)
	}
	if x.Config.VTunSettings.Compress {
		b = snappy.Encode(nil, b)
	}
	return b, nil
}

func (x *Server) BasicDecode(b []byte) ([]byte, error) {
	var err error
	if x.Config.VTunSettings.Compress {
		b, err = snappy.Decode(nil, b)
		if err != nil {
			return nil, err
		}
	}
	if x.Config.VTunSettings.Obfs {
		b = cipher.XOR(b)
	}
	return b, nil
}

func (x *Server) ExtendEncode(b []byte) ([]byte, error) {
	var err error
	if x.Config.VTunSettings.Obfs {
		b = cipher.XOR(b)
	}
	b, err = x.xp.Encode(b)
	if err != nil {
		return nil, err
	}
	if x.Config.VTunSettings.Compress {
		b = snappy.Encode(nil, b)
	}
	return b, nil
}

func (x *Server) ExtendDecode(b []byte) ([]byte, error) {
	var err error
	if x.Config.VTunSettings.Compress {
		b, err = snappy.Decode(nil, b)
		if err != nil {
			return nil, err
		}
	}
	b, err = x.xp.Decode(b)
	if err != nil {
		return nil, err
	}
	if x.Config.VTunSettings.Obfs {
		b = cipher.XOR(b)
	}
	return b, nil
}

func (x *Server) AuthKey() *xproto.AuthKey {
	return x.authKey
}

func GetTimeout() time.Time {
	return time.Now().Add(time.Second * 9)
}
