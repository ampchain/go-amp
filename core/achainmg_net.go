package AmpChaincore

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/ampchain/go-amp/common/config"
	"github.com/ampchain/go-amp/global"
	"github.com/ampchain/go-amp/p2pv2"
	amper_p2p "github.com/ampchain/go-amp/p2pv2/pb"
	"github.com/ampchain/go-amp/pb"
)

// RegisterSubscriber register p2pv2 msg type
func (xm *AChainMG) RegisterSubscriber() error {
	xm.Log.Trace("Start to Register Subscriber")
	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(xm.msgChan, amper_p2p.AmperMessage_POSTTX, nil, "")); err != nil {
		return err
	}

	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(xm.msgChan, amper_p2p.AmperMessage_SENDBLOCK, nil, "")); err != nil {
		return err
	}

	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(xm.msgChan, amper_p2p.AmperMessage_BATCHPOSTTX, nil, "")); err != nil {
		return err
	}

	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(nil, amper_p2p.AmperMessage_GET_BLOCK, xm.handleGetBlock, "")); err != nil {
		return err
	}

	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(nil, amper_p2p.AmperMessage_GET_BLOCKCHAINSTATUS, xm.handleGetBlockChainStatus, "")); err != nil {
		return err
	}

	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(nil, amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS, xm.handleConfirmBlockChainStatus, "")); err != nil {
		return err
	}

	if _, err := xm.P2pv2.Register(p2pv2.NewSubscriber(nil, amper_p2p.AmperMessage_GET_RPC_PORT, xm.handleGetRPCPort, "")); err != nil {
		return err
	}

	xm.Log.Trace("Stop to Register Subscriber")
	return nil
}

// StartLoop dispatch msg received
func (xm *AChainMG) StartLoop() {
	xm.Log.Info("AChainMg start loop to process net msg")
	for {
		select {
		case msg := <-xm.msgChan:
			// handle received msg
			xm.Log.Info("AChainMG get msg", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
			go xm.handleReceivedMsg(msg)
		}
	}
}

func (xm *AChainMG) handleReceivedMsg(msg *amper_p2p.AmperMessage) {
	// check msg type
	msgType := msg.GetHeader().GetType()
	if msgType != amper_p2p.AmperMessage_POSTTX && msgType != amper_p2p.AmperMessage_SENDBLOCK && msgType !=
		amper_p2p.AmperMessage_BATCHPOSTTX {
		xm.Log.Warn("Received msg cannot handled!", "logid", msg.GetHeader().GetLogid())
		return
	}

	// verify msg
	if !amper_p2p.VerifyDataCheckSum(msg) {
		xm.Log.Warn("Verify Data error!", "logid", msg.GetHeader().GetLogid())
		return
	}

	// process msg
	switch msgType {
	case amper_p2p.AmperMessage_POSTTX:
		xm.handlePostTx(msg)
	case amper_p2p.AmperMessage_SENDBLOCK:
		xm.HandleSendBlock(msg)
	case amper_p2p.AmperMessage_BATCHPOSTTX:
		xm.handleBatchPostTx(msg)
	}
}

func (xm *AChainMG) handlePostTx(msg *amper_p2p.AmperMessage) {
	txStatus := &pb.TxStatus{}

	// Unmarshal msg
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), txStatus)
	if err != nil {
		xm.Log.Error("handlePostTx Unmarshal msg to tx error", "logid", msg.GetHeader().GetLogid())
		return
	}

	// process tx
	if txStatus.Header == nil {
		txStatus.Header = global.GHeader()
	}
	if _, needRepost, _ := xm.ProcessTx(txStatus); needRepost {
		opts := []p2pv2.MessageOption{
			p2pv2.WithFilters([]p2pv2.FilterStrategy{p2pv2.DefaultStrategy}),
			p2pv2.WithBcName(msg.GetHeader().GetBcname()),
		}
		go xm.P2pv2.SendMessage(context.Background(), msg, opts...)
	}
	return
}

// ProcessTx process tx, move from server/server.go
func (xm *AChainMG) ProcessTx(in *pb.TxStatus) (*pb.CommonReply, bool, error) {
	out := &pb.CommonReply{Header: &pb.Header{Logid: in.Header.Logid}}

	if err := validatePostTx(in); err != nil {
		out.Header.Error = pb.AChainErrorEnum_VALIDATE_ERROR
		xm.Log.Trace("PostTx validate param errror", "logid", in.Header.Logid, "error", err.Error())
		return out, false, err
	}

	if len(in.Tx.TxInputs) == 0 && !xm.Cfg.Utxo.NonUtxo {
		out.Header.Error = pb.AChainErrorEnum_CONNECT_REFUSE // 拒绝
		xm.Log.Warn("PostTx TxInputs can not be null while need utxo!", "logid", in.Header.Logid)
		return out, false, nil
	}

	bc := xm.Get(in.Bcname)
	if bc == nil {
		out.Header.Error = pb.AChainErrorEnum_CONNECT_REFUSE // 拒绝
		return out, false, nil
	}

	if xm.Cfg.FeeConfig.NeedFee {
		fee := in.Tx.GetFee()
		txSize := int64(len(in.Tx.Desc))
		size := int64(0)
		if txSize%1024 == 0 {
			size = txSize / 1024
		} else {
			size = txSize/1024 + 1
		}
		cost := big.NewInt((size) * xm.Cfg.FeeConfig.UnitFee)
		if res := fee.Cmp(cost); res < 0 {
			out.Header.Error = pb.AChainErrorEnum_TX_FEE_NOT_ENOUGH_ERROR
			xm.Log.Warn("PostTx fee not enough for storage!", "logid", in.Header.Logid)
			return out, false, nil
		}
	}

	hd := &global.XContext{Timer: global.NewXTimer()}

	if bc.GetNodeMode() == config.NodeModeFastSync {
		out.Header.Error = pb.AChainErrorEnum_CONNECT_REFUSE // 拒绝
		xm.Log.Warn("PostTx NodeMode is FAST_SYNC, refused!")
		return out, false, nil
	}
	out, needRepost := bc.PostTx(in, hd)
	return out, needRepost, nil
}

// HandleSendBlock handle SENDBLOCK type msg
func (xm *AChainMG) HandleSendBlock(msg *amper_p2p.AmperMessage) {
	block := &pb.Block{}
	xm.Log.Trace("Start to HandleSendBlock", "logid", msg.GetHeader().GetLogid(), "checksum", msg.GetHeader().GetDataCheckSum())
	// Unmarshal msg
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), block)
	if err != nil {
		xm.Log.Error("HandleSendBlock Unmarshal msg to block error", "logid", msg.GetHeader().GetLogid())
		return
	}
	// process block
	if block.Header == nil {
		block.Header = global.GHeader()
	}
	if err := xm.ProcessBlock(block); err != nil {
		xm.Log.Error("HandleSendBlock ProcessBlock error", "error", err.Error())
		return
	}
	bcname := block.GetBcname()
	bc := xm.Get(bcname)
	filters := []p2pv2.FilterStrategy{p2pv2.DefaultStrategy}
	if bc.NeedCoreConnection() {
		filters = append(filters, p2pv2.CorePeersStrategy)
	}
	opts := []p2pv2.MessageOption{
		p2pv2.WithFilters(filters),
		p2pv2.WithBcName(bcname),
	}
	go xm.P2pv2.SendMessage(context.Background(), msg, opts...)
	return
}

// ProcessBlock process block
func (xm *AChainMG) ProcessBlock(block *pb.Block) error {
	if err := validateSendBlock(block); err != nil {
		xm.Log.Error("ProcessBlock validateSendBlock error", "error", err.Error())
		return err
	}

	xm.Log.Trace("Start to dealwith SendBlock", "blockid", global.F(block.GetBlockid()))
	bc := xm.Get(block.GetBcname())
	if bc == nil {
		xm.Log.Error("ProcessBlock error", "error", "bc not exist")
		return ErrBlockChainNotExist
	}
	hd := &global.XContext{Timer: global.NewXTimer()}
	if err := bc.SendBlock(block, hd); err != nil {
		xm.Log.Error("ProcessBlock SendBlock error", "err", err)
		return err
	}
	meta := bc.Ledger.GetMeta()
	xm.Log.Info("SendBlock", "cost", hd.Timer.Print(), "genesis", fmt.Sprintf("%x", meta.RootBlockid),
		"last", fmt.Sprintf("%x", meta.TipBlockid),
		"height", meta.TrunkHeight, "utxo", global.F(bc.Utxovm.GetLatestBlockid()))
	return nil
}

func (xm *AChainMG) handleBatchPostTx(msg *amper_p2p.AmperMessage) {
	batchTxs := &pb.BatchTxs{}
	// Unmarshal msg
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), batchTxs)
	if err != nil {
		xm.Log.Error("handleBatchPostTx Unmarshal msg to BatchTxs error", "logid", msg.GetHeader().GetLogid())
		return
	}

	// process batch post tx
	txs, err := xm.ProcessBatchTx(batchTxs)
	if err != nil {
		xm.Log.Error("HandleSendBlock ProcessBlock error", "error", err.Error())
		return
	}
	if len(txs.Txs) != 0 {
		txsData, err := proto.Marshal(txs)
		if err != nil {
			xm.Log.Error("handleBatchPostTx Marshal txs error", "error", err)
			return
		}
		msg.Data.MsgInfo = txsData
		msg.Header.DataCheckSum = amper_p2p.CalDataCheckSum(msg)
		opts := []p2pv2.MessageOption{
			p2pv2.WithFilters([]p2pv2.FilterStrategy{p2pv2.DefaultStrategy}),
			p2pv2.WithBcName(msg.GetHeader().GetBcname()),
		}
		go xm.P2pv2.SendMessage(context.Background(), msg, opts...)
	}
	return
}

// ProcessBatchTx process batch tx
func (xm *AChainMG) ProcessBatchTx(batchTx *pb.BatchTxs) (*pb.BatchTxs, error) {
	succTxs := []*pb.TxStatus{}
	for _, v := range batchTx.Txs {
		_, needRepost, _ := xm.ProcessTx(v)
		if needRepost {
			succTxs = append(succTxs, v)
		}
	}
	batchTx.Txs = succTxs
	return batchTx, nil
}

// 处理getBlock消息回调函数
func (xm *AChainMG) handleGetBlock(ctx context.Context, msg *amper_p2p.AmperMessage) (*amper_p2p.AmperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	logid := msg.GetHeader().GetLogid()
	xm.Log.Trace("Start to handleGetBlock", "bcname", bcname, "logid", logid)
	block := &pb.Block{Header: global.GHeader()}
	if !amper_p2p.VerifyDataCheckSum(msg) {
		xm.Log.Warn("handleGetBlock verify msg error", "log_id", logid)
		resBuf, _ := proto.Marshal(block)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCK_RES, resBuf, amper_p2p.AmperMessage_CHECK_SUM_ERROR)
		return res, errors.New("verify msg error")
	}
	bid := &pb.BlockID{}
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), bid)

	if err != nil {
		xm.Log.Error("handleGetBlock unmarshal msg error", "error", err.Error())
		resBuf, _ := proto.Marshal(block)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCK_RES, resBuf, amper_p2p.AmperMessage_UNMARSHAL_MSG_BODY_ERROR)
		return res, errors.New("unmarshal msg error")
	}

	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Error("handleGetBlock Get blockchain error", "error", "blockchain not exit", "bcname", bcname)
		resBuf, _ := proto.Marshal(block)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCK_RES, resBuf, amper_p2p.AmperMessage_BLOCKCHAIN_NOTEXIST)
		return res, errors.New("blockChain not exit")
	}
	block = bc.GetBlock(bid)
	xm.Log.Trace("Start to dealwith GetBlock result", "logid", logid,
		"blockid", block.GetBlock().GetBlockid(), "height", block.GetBlock().GetHeight())
	if block.GetHeader().GetError() != pb.AChainErrorEnum_SUCCESS {
		xm.Log.Error("handleGetBlock GetBlock error", "error", block.GetHeader().GetError())
		resBuf, _ := proto.Marshal(block)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCK_RES, resBuf, amper_p2p.AmperMessage_GET_BLOCK_ERROR)
		return res, errors.New("getBlock error")
	}

	resBuf, _ := proto.Marshal(block)
	res, err := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
		amper_p2p.AmperMessage_GET_BLOCK_RES, resBuf, amper_p2p.AmperMessage_SUCCESS)
	return res, err
}

// 处理getBlockChainStatus消息回调函数
func (xm *AChainMG) handleGetBlockChainStatus(ctx context.Context, msg *amper_p2p.AmperMessage) (*amper_p2p.AmperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	logid := msg.GetHeader().GetLogid()
	xm.Log.Trace("Start to handleGetBlockChainStatus", "bcname", bcname, "logid", logid)
	if !amper_p2p.VerifyDataCheckSum(msg) {
		xm.Log.Warn("handleGetBlockChainStatus verify msg error", "log_id", logid)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_CHECK_SUM_ERROR)
		return res, errors.New("verify msg error")
	}
	bcStatus := &pb.BCStatus{}
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), bcStatus)
	if err != nil {
		xm.Log.Error("handleGetBlockChainStatus unmarshal msg error", "error", err.Error())
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_UNMARSHAL_MSG_BODY_ERROR)
		return res, errors.New("unmarshal msg error")
	}
	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Error("handleGetBlockChainStatus Get blockchain error", "error", "blockchain not exit", "bcname", bcname)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_BLOCKCHAIN_NOTEXIST)
		return res, errors.New("blockChain not exit")
	}
	bcStatusRes := bc.GetBlockChainStatus(bcStatus)
	if bcStatusRes.GetHeader().GetError() != pb.AChainErrorEnum_SUCCESS {
		xm.Log.Error("handleGetBlockChainStatus Get blockchain error", "error", bcStatusRes.GetHeader().GetError())
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_GET_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_GET_BLOCKCHAIN_ERROR)
		return res, errors.New("get BlockChainStatus error")
	}
	resBuf, _ := proto.Marshal(bcStatusRes)
	res, err := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
		amper_p2p.AmperMessage_GET_BLOCKCHAINSTATUS_RES, resBuf, amper_p2p.AmperMessage_SUCCESS)
	return res, err
}

// 处理confirm blockChain status 回调函数
func (xm *AChainMG) handleConfirmBlockChainStatus(ctx context.Context, msg *amper_p2p.AmperMessage) (*amper_p2p.AmperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	logid := msg.GetHeader().GetLogid()
	xm.Log.Trace("Start to handleConfirmBlockChainStatus", "bcname", bcname, "logid", logid)
	if !amper_p2p.VerifyDataCheckSum(msg) {
		xm.Log.Warn("handleConfirmBlockChainStatus verify msg error", "log_id", logid)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_CHECK_SUM_ERROR)
		return res, errors.New("verify msg error")
	}

	bcStatus := &pb.BCStatus{}
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), bcStatus)
	if err != nil {
		xm.Log.Error("handleConfirmBlockChainStatus unmarshal msg error", "error", err.Error())
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_UNMARSHAL_MSG_BODY_ERROR)
		return res, errors.New("unmarshal msg error")
	}

	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Error("handleConfirmBlockChainStatus Get blockchain error", "error", "blockchain not exit", "bcname", bcname)
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_BLOCKCHAIN_NOTEXIST)
		return res, errors.New("blockChain not exit")
	}
	tipStatus := bc.ConfirmTipBlockChainStatus(bcStatus)
	if tipStatus.GetHeader().GetError() != pb.AChainErrorEnum_SUCCESS {
		xm.Log.Error("handleConfirmBlockChainStatus ConfirmTipBlockChainStatus error", "error", tipStatus.GetHeader().GetError())
		res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
			amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS_ERROR)
		return res, errors.New("confirmBlockChainStatus error")
	}
	resBuf, _ := proto.Marshal(tipStatus)
	res, err := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, bcname, logid,
		amper_p2p.AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, resBuf, amper_p2p.AmperMessage_SUCCESS)
	return res, err
}

// 处理获取RPC端口回调函数
func (xm *AChainMG) handleGetRPCPort(ctx context.Context, msg *amper_p2p.AmperMessage) (*amper_p2p.AmperMessage, error) {
	xm.Log.Trace("Start to handleGetRPCPort", "logid", msg.GetHeader().GetLogid())
	_, port, err := net.SplitHostPort(xm.Cfg.TCPServer.Port)
	if err != nil {
		xm.Log.Error("handleGetRPCPort SplitHostPort error", "error", err.Error())
		return nil, err
	}
	return amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, "", msg.GetHeader().GetLogid(), amper_p2p.AmperMessage_GET_RPC_PORT_RES, []byte(":"+port), amper_p2p.AmperMessage_NONE)
}
