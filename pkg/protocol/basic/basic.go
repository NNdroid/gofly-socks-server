package basic

import (
	"context"
	"github.com/patrickmn/go-cache"
	"gofly/pkg/config"
	"gofly/pkg/statistics"
	"time"
)

type ServerForApi interface {
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
}

func (x *Server) ConvertDstAddr(packet []byte) {
	//
}

func (x *Server) ConvertSrcAddr(packet []byte) {
	//
}

func GetTimeout() time.Time {
	return time.Now().Add(time.Second * 3)
}
