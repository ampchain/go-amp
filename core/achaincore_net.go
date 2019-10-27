package AmpChaincore

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/ampchain/go-amp/p2pv2"
	amper_p2p "github.com/ampchain/go-amp/p2pv2/pb"
	"github.com/ampchain/go-amp/pb"
)

// BroadCastGetBlock get block from p2p network nodes
func (xc *AChainCore) BroadCastGetBlock(bid *pb.BlockID) *pb.Block {
	msgbuf, err := proto.Marshal(bid)
	if err != nil {
		xc.log.Warn("BroadCastGetBlock Marshal msg error", "error", err)
		return nil
	}
	msg, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bid.GetBcname(), "", amper_p2p.AmperMessage_GET_BLOCK, msgbuf, amper_p2p.AmperMessage_NONE)
	filters := []p2pv2.FilterStrategy{p2pv2.NearestBucketStrategy}
	if xc.NeedCoreConnection() {
		filters = append(filters, p2pv2.CorePeersStrategy)
	}
	opts := []p2pv2.MessageOption{
		p2pv2.WithFilters(filters),
		p2pv2.WithBcName(xc.bcname),
	}
	res, err := xc.P2pv2.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil || len(res) < 1 {
		return nil
	}

	for _, v := range res {
		if v.GetHeader().GetErrorType() != amper_p2p.AmperMessage_SUCCESS {
			continue
		}

		block := &pb.Block{}
		err = proto.Unmarshal(v.GetData().GetMsgInfo(), block)
		if err != nil {
			xc.log.Warn("BroadCastGetBlock unmarshal error", "error", err)
			continue
		} else {
			if block.Block == nil {
				continue
			}
			return block
		}
	}
	return nil
}
