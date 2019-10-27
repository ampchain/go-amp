/*
 * 
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/syndtr/goleveldb/leveldb/errors"

	"github.com/ampchain/go-amp/common/config"
	"github.com/ampchain/go-amp/common/log"
	AmpChaincore "github.com/ampchain/go-amp/core"
	"github.com/ampchain/go-amp/p2pv2"
	"github.com/ampchain/go-amp/server"
)

var (
	buildVersion = ""
	commitHash   = ""
	buildDate    = ""
)

// Start init and star chain node
func Start(cfg *config.NodeConfig) error {
	xlog, err := log.OpenDefaultLog(&cfg.Log)
	if err != nil {
		err := errors.New("open log fail")
		return err
	}
	xlog.Info("debug info", "root host", cfg.ConsoleConfig.Host)

	// start node
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	cfg.VisitAll()
	xlog.Trace("Hello BlockChain")

	// 注册优雅关停信号, 包括ctrl + C 和 kill 信号
	xlog.Trace("register stopping handler")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigc)

	// Init p2pv2
	p2pV2Serv, err := p2pv2.NewP2PServerV2(cfg.P2pV2, xlog)
	if err != nil {
		panic(err)
	}

	xcmg := AmpChaincore.AChainMG{}
	if err = xcmg.Init(xlog, cfg, p2pV2Serv); err != nil {
		panic(err)
	}

	if cfg.CPUProfile != "" {
		perfFile, perr := os.Create(cfg.CPUProfile)
		if perr != nil {
			panic(perr)
		}
		pprof.StartCPUProfile(perfFile)
	}
	// 启动挖矿结点
	xcmg.Start()
	go server.SerRun(&xcmg)
	for {
		select {
		case <-sigc:
			xlog.Info("Got terminate, start to shutting down, please wait...")
			close(xcmg.Quit)
		case <-xcmg.Quit:
			xlog.Info("Got xcmg quit, start to shutting down, please wait...")
			Stop(&xcmg)
			return nil
		}
	}
}

// Stop gracefully shut down, 各个模块实现自己需要优雅关闭的资源并在此处调用即可
func Stop(AmpChainmg *AmpChaincore.AChainMG) {
	if AmpChainmg.Cfg.CPUProfile != "" {
		pprof.StopCPUProfile()
	}
	if AmpChainmg.Cfg.MemProfile != "" {
		f, err := os.Create(AmpChainmg.Cfg.MemProfile)
		if err == nil {
			pprof.WriteHeapProfile(f)
			f.Close()
		}
	}
	AmpChainmg.Stop()
	AmpChainmg.Log.Info("All modules have stopped!")
	pprof.StopCPUProfile()
	return
}

func main() {
	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	var showVersion bool
	flags.BoolVar(&showVersion, "version", false, "show AmpChain version")

	cfg := config.NewNodeConfig()
	cfg.LoadConfig()
	cfg.ApplyFlags(flags)

	flags.Parse(os.Args[1:])

	if showVersion {
		fmt.Printf("%s-%s %s\n", buildVersion, commitHash, buildDate)
		return
	}

	err := Start(cfg)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(-1)
	}
}
