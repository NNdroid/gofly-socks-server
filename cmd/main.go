package main

import (
	"flag"
	"go.uber.org/zap"
	"gofly"
	"gofly/pkg/config"
	"gofly/pkg/engine"
	"gofly/pkg/logger"
	"gofly/pkg/utils"
	"log"
	"os"
)

var (
	_configFilePath string
	_flagQuiet      bool
	_config         *config.Config
	_flagVersion    bool
)

var (
	_version   = "v1.0.20230731"
	_gitHash   = "nil"
	_buildTime = "nil"
	_goVersion = "nil"
)

func displayVersionInfo() {
	log.Printf("version %s", _version)
	log.Printf("git hash %s", _gitHash)
	log.Printf("build time %s", _buildTime)
	log.Printf("go version %s", _goVersion)
}

func init() {
	log.Printf("\n  ____       _____ _            ____             _        \n / ___| ___ |  ___| |_   _     / ___|  ___   ___| | _____ \n| |  _ / _ \\| |_  | | | | |____\\___ \\ / _ \\ / __| |/ / __|\n| |_| | (_) |  _| | | |_| |_____|__) | (_) | (__|   <\\__ \\\n \\____|\\___/|_|   |_|\\__, |    |____/ \\___/ \\___|_|\\_\\___/\n                     |___/                                ")
	logger.Init()
	flag.StringVar(&_configFilePath, "c", "config.yaml", "the path of configuration file")
	flag.BoolVar(&_flagQuiet, "quiet", false, "quiet for log print.")
	flag.BoolVar(&_flagVersion, "v", false, "print version info.")
	flag.Parse()
	if _flagVersion {
		displayVersionInfo()
		os.Exit(0)
	}
	if _flagQuiet {
		logger.Cfg.Level.SetLevel(zap.ErrorLevel)
		engine.SetLogLevel(false)
	}
	if !utils.IsFile(_configFilePath) || !utils.ExistsFile(_configFilePath) {
		logger.Logger.Fatal("configure file not found!")
	}
	dat, err := utils.ReadFile(_configFilePath)
	if err != nil {
		logger.Logger.Fatal("read configure file fail!", zap.Error(err))
	}
	_config, err = config.Parse(dat)
	if err != nil {
		logger.Logger.Fatal("parse configure file fail!", zap.Error(err))
	}
	if _flagQuiet {
		_config.VTunSettings.Verbose = false
	}
}

func main() {
	gofly.StartServer(_config)
}
