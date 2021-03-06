package p2pv2

import (
	"context"

	p2pPb "github.com/ampchain/go-amp/p2pv2/pb"
)

// CorePeersInfo defines the peers' info for core nodes
// By setting this info, we can keep some core peers always connected directly
// It's useful for keeping DPoS key network security and for some BFT-like consensus
type CorePeersInfo struct {
	Name           string   // distinguished name of the node routing
	CurrentTermNum int64    // the current term number
	CurrentPeerIDs []string // current core peer IDs
	NextPeerIDs    []string // upcoming core peer IDs
}

// P2PServer is the p2p server interface of Amper
type P2PServer interface {
	Start()
	Stop()

	// 注册订阅者，支持多个用户订阅同一类消息
	Register(sub *Subscriber) (*Subscriber, error)
	// 注销订阅者，需要根据当时注册时返回的Subscriber实例删除
	UnRegister(sub *Subscriber) error

	SendMessage(context.Context, *p2pPb.AmperMessage, ...MessageOption) error
	// todo: 将请求的参数改为Option的方式
	SendMessageWithResponse(context.Context, *p2pPb.AmperMessage, ...MessageOption) ([]*p2pPb.AmperMessage, error)

	GetNetURL() string
	// 查询所连接节点的信息
	GetPeerUrls() []string

	// SetCorePeers set core peers' info to P2P server
	SetCorePeers(cp *CorePeersInfo) error

	// SetAChainAddr Set AmpChain address from AmpChaincore
	SetAChainAddr(bcname string, info *AChainAddrInfo)
}
