package pluginmgr

import (
	"github.com/ampchain/go-amp/common/config"
	"github.com/ampchain/go-amp/common/log"

	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// AChainPM is the plugin manager wrapper for AmpChain
type AChainPM struct {
	PluginMgr *PluginMgr
	xlog      log.Logger
}

var xpm *AChainPM
var locker sync.Mutex

// Init is not necessary, there is a default value
func Init(nc *config.NodeConfig) error {
	if xpm == nil {
		locker.Lock()
		defer locker.Unlock()

		if xpm == nil {
			confPath := nc.PluginConfPath
			logConf := nc.Log
			logConf.Filename += "_pm"
			return createAChainPM(getAChainRoot(), confPath, &logConf)
		}
	}
	return nil
}

func createAChainPM(rootFolder string, confPath string, logConf *config.LogConfig) error {
	logger, err := log.OpenLog(logConf)
	if err != nil {
		fmt.Println("Init pluginmgr log failed!")
		return err
	}

	pluginMgr, err := CreateMgr(rootFolder, confPath, logger)
	if err != nil {
		logger.Warn("Init pluginmgr failed!")
		return err
	}

	xpm = &AChainPM{
		xlog:      logger,
		PluginMgr: pluginMgr,
	}

	return nil
}

func createDefaultXchianPM() error {
	pluginConf := "./conf/plugins.conf"
	logFolder, err := makeFullPath("logs")
	if err != nil {
		logFolder = "./logs"
	}

	logConfig := &config.LogConfig{
		Module:         "pluginmgr",
		Filepath:       logFolder,
		Filename:       "pluginmgr",
		Fmt:            "logfmt",
		Console:        false,
		Level:          "trace",
		Async:          false,
		RotateInterval: 60 * 24, // rotate every 1 day
		RotateBackups:  7,       // keep old log files for 7 days
	}

	return createAChainPM(getAChainRoot(), pluginConf, logConfig)
}

// GetPluginMgr return plugin manager instance
func GetPluginMgr() (*AChainPM, error) {
	// if not initialized AChainPM, use default value to init
	if xpm == nil {
		err := createDefaultXchianPM()
		return xpm, err
	}

	return xpm, nil
}

func makeFullPath(relativePath string) (string, error) {
	AmpChainRoot := getAChainRoot()
	if AmpChainRoot != "" {
		return path.Join(AmpChainRoot, relativePath), nil
	}

	return filepath.Abs(relativePath)
}

func getAChainRoot() string {
	return os.Getenv("AChain_ROOT")
}
