package gofly

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gofly/pkg/config"
	"gofly/pkg/device/tun"
	"gofly/pkg/engine"
	"gofly/pkg/logger"
	"gofly/pkg/protocol/basic"
	"gofly/pkg/protocol/reality"
	"gofly/pkg/protocol/ws"
	"gofly/pkg/statistics"
	"log"
)

var _ctx context.Context
var cancel context.CancelFunc
var stats *statistics.Statistics

func StartServer(config *config.Config) {
	_ctx, cancel = context.WithCancel(context.Background())
	stats = &statistics.Statistics{}
	go stats.AutoUpdateChartData()
	bs := basic.Server{
		Config:     config,
		ReadFunc:   ReadFromTun,
		WriteFunc:  WriteToTun,
		CTX:        _ctx,
		Statistics: stats,
	}
	go RunTun2Socks(config, _ctx)
	var server basic.ServerForApi
	var err error
	switch config.VTunSettings.Protocol {
	case "ws", "wss":
		err = config.WebSocketSettings.Check()
		if err != nil {
			logger.Logger.Sugar().Errorf("error: %v\n", zap.Error(err))
			return
		}
		server = &ws.Server{
			Server: bs,
		}
		break
	case "reality":
		err = config.RealitySettings.Check()
		if err != nil {
			logger.Logger.Sugar().Errorf("error: %v\n", zap.Error(err))
			return
		}
		server = &reality.Server{
			Server: bs,
		}
		break
	default:
		log.Panic(errors.New("unsupported protocol"))
	}
	//init server
	server.Init()
	//start server
	server.StartServerForApi()
}

func RunTun2Socks(config *config.Config, _ctx context.Context) {
	engine.Insert(&config.Tun2SocksSettings)
	engine.Start()
	defer engine.Stop()
	<-_ctx.Done()
}

func ReadFromTun(bts []byte) (int, error) {
	return tun.ReadFromTun(bts)
}

func WriteToTun(bts []byte) int {
	n, _ := tun.WriteToTun(bts)
	return n
}

func Close() {
	cancel()
}
