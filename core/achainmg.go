package AmpChaincore

import (
	"errors"
	"io/ioutil"
	"sync"

	"github.com/ampchain/log15"
	"github.com/ampchain/go-amp/common/config"
	"github.com/ampchain/go-amp/common/probe"
	"github.com/ampchain/go-amp/contract/kernel"
	"github.com/ampchain/go-amp/p2pv2"
	amper_p2p "github.com/ampchain/go-amp/p2pv2/pb"
)

// AChainMG manage all chains
type AChainMG struct {
	Log   log.Logger
	Cfg   *config.NodeConfig
	P2pv2 p2pv2.P2PServer
	// msgChan is the message subscribe from net
	msgChan    chan *amper_p2p.AmperMessage
	chains     *sync.Map
	rootKernel *kernel.Kernel
	datapath   string
	Ukeys      *sync.Map //address -> scrkey
	Speed      *probe.SpeedCalc
	Quit       chan struct{}
	nodeMode   string
}

// Init init instance of AChainMG
func (xm *AChainMG) Init(log log.Logger, cfg *config.NodeConfig,
	p2pV2 p2pv2.P2PServer) error {
	xm.Log = log
	xm.chains = new(sync.Map)
	xm.datapath = cfg.Datapath
	xm.Cfg = cfg
	xm.P2pv2 = p2pV2
	xm.msgChan = make(chan *amper_p2p.AmperMessage, p2pv2.MsgChanSize)

	xm.Speed = probe.NewSpeedCalc("sum")
	xm.Quit = make(chan struct{})
	xm.nodeMode = cfg.NodeMode

	dir, err := ioutil.ReadDir(xm.datapath)
	if err != nil {
		xm.Log.Error("can't open data", "datapath", xm.datapath)
		return err
	}
	for _, fi := range dir {
		if fi.IsDir() { // 忽略非目录
			xm.Log.Trace("--------find " + fi.Name())
			aKernel := &kernel.Kernel{}
			aKernel.Init(xm.datapath, xm.Log, xm, fi.Name())
			x := &AChainCore{}
			err := x.Init(fi.Name(), log, cfg, p2pV2, aKernel, xm.nodeMode)
			if err != nil {
				return err
			}
			if fi.Name() == "amper" {
				xm.rootKernel = aKernel
			}
			xm.chains.Store(fi.Name(), x)
		}
	}
	if xm.rootKernel == nil {
		err := errors.New("amper chain not found")
		xm.Log.Error("can not find amper chain, please create it first", "err", err)
		return err
	}
	xm.rootKernel.SetNewChainWhiteList(cfg.Kernel.NewChainWhiteList)
	xm.rootKernel.SetMinNewChainAmount(cfg.Kernel.MinNewChainAmount)
	/*for _, x := range xm.chains {
		go x.SyncBlocks()
	}*/
	if err := xm.RegisterSubscriber(); err != nil {
		return err
	}
	go xm.Speed.ShowLoop(xm.Log)
	return nil
}

// Get return specific instance of blockchain by blockchain name from map
func (xm *AChainMG) Get(name string) *AChainCore {
	v, ok := xm.chains.Load(name)
	if ok {
		xc := v.(*AChainCore)
		return xc
	}
	return nil
}

// Set put <blockname, blockchain instance> into map
func (xm *AChainMG) Set(name string, xc *AChainCore) {
	xm.chains.Store(name, xc)
}

// GetAll returns all blockchains name
func (xm *AChainMG) GetAll() []string {
	var bcs []string
	xm.chains.Range(func(k, v interface{}) bool {
		xc := v.(*AChainCore)
		bcs = append(bcs, xc.bcname)
		return true
	})
	return bcs
}

// Start start all blockchain instances
func (xm *AChainMG) Start() {
	xm.chains.Range(func(k, v interface{}) bool {
		xc := v.(*AChainCore)
		xm.Log.Trace("start chain " + k.(string))
		go xc.Miner()
		return true
	})
	go xm.StartLoop()
}

// Stop stop all blockchain instances
func (xm *AChainMG) Stop() {
	xm.chains.Range(func(k, v interface{}) bool {
		xc := v.(*AChainCore)
		xm.Log.Trace("stop chain " + k.(string))
		xc.Stop()
		return true
	})
	if xm.P2pv2 != nil {
		xm.P2pv2.Stop()
	}
}

// CreateBlockChain create an instance of blockchain
func (xm *AChainMG) CreateBlockChain(name string, data []byte) (*AChainCore, error) {
	if _, ok := xm.chains.Load(name); ok {
		xm.Log.Warn("chains[" + name + "] is exist")
		return nil, ErrBlockChainIsExist
	}

	if err := xm.rootKernel.CreateBlockChain(name, data); err != nil {
		return nil, err
	}
	return xm.addBlockChain(name)
}

func (xm *AChainMG) addBlockChain(name string) (*AChainCore, error) {
	x := &AChainCore{}
	aKernel := xm.rootKernel
	if name != "amper" {
		aKernel = &kernel.Kernel{}
		aKernel.Init(xm.datapath, xm.Log, xm, name)
	}
	err := x.Init(name, xm.Log, xm.Cfg, xm.P2pv2, aKernel, xm.nodeMode)
	if err != nil {
		xm.Log.Warn("AChainCore init error")
		xm.rootKernel.RemoveBlockChainData(name)
		return nil, err
	}
	xm.Set(name, x)
	return x, nil
}

// RegisterBlockChain load an instance of blockchain and start it dynamically
func (xm *AChainMG) RegisterBlockChain(name string) error {
	xc, err := xm.addBlockChain(name)
	if err != nil {
		return err
	}
	go xc.Miner()
	return err
}

// UnloadBlockChain unload an instance of blockchain and stop it dynamically
func (xm *AChainMG) UnloadBlockChain(name string) error {
	v, ok := xm.chains.Load(name)
	if !ok {
		return ErrBlockChainIsExist
	}
	xm.chains.Delete(name) //从AmpChainmg的map里面删了，就不会收到新的请求了
	//然后停止这个链
	xc := v.(*AChainCore)
	xc.Stop()
	return nil
}
