package p2pv2

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/ampchain/go-amp/common/config"
	crypto_client "github.com/ampchain/go-amp/crypto/client"
	"github.com/ampchain/go-amp/crypto/hash"
	amper_p2p "github.com/ampchain/go-amp/p2pv2/pb"
	"github.com/ampchain/go-amp/pb"
)

// RegisterSubsriber register handleMessage callback fucntion
func (p *P2PServerV2) registerSubscriber() error {
	if _, err := p.Register(NewSubscriber(nil, amper_p2p.AmperMessage_GET_AUTHENTICATION,
		p.handleGetAuthentication, "")); err != nil {
		return err
	}

	p.log.Trace("Stop to Register Subscriber")
	return nil
}

// handleGetAuthentication callback function for handling identity authentication
func (p *P2PServerV2) handleGetAuthentication(ctx context.Context, msg *amper_p2p.AmperMessage) (*amper_p2p.AmperMessage, error) {
	logid := msg.Header.Logid
	auths := &pb.IdentityAuths{}
	errRes := errorHandleGetAuthenMsg(logid)
	err := proto.Unmarshal(msg.Data.MsgInfo, auths)
	if err != nil {
		p.log.Error("handleGetAuthentication unmarshal msg error", "error", err.Error())
		return errRes, errors.New("unmarshal msg error")
	}
	p.log.Trace("Start to handleGetAuthentication", "logid", logid, "authsrequest", auths)

	addrs := make([]string, 0, len(auths.Auth))
	s := ctx.Value("Stream").(*Stream)
	for _, v := range auths.Auth {
		if s.p.Pretty() != v.PeerID {
			p.log.Error("handleGetAuthentication peerID inconsistency", "s.PeerID", s.p.Pretty(), "v.PeerID", v.PeerID)
			return errRes, errors.New("handleGetAuthentication peerID inconsistency")
		}

		cryptoClient, err := crypto_client.CreateCryptoClientFromJSONPublicKey(v.Pubkey)
		if err != nil {
			p.log.Error("handleGetAuthentication Create crypto client error", "error", err.Error())
			return errRes, errors.New("handleGetAuthentication Create crypto client error")
		}

		publicKey, err := cryptoClient.GetEcdsaPublicKeyFromJSON(v.Pubkey)
		if err != nil {
			p.log.Error("handleGetAuthentication GetEcdsaPublicKeyFromJSON error", "error", err.Error())
			return errRes, err
		}

		isMatch, _ := cryptoClient.VerifyAddressUsingPublicKey(v.Addr, publicKey)
		if !isMatch {
			p.log.Error("handleGetAuthentication address and public key not match")
			return errRes, errors.New("handleGetAuthentication address and public key not match")
		}

		tsNow := time.Now().Unix()
		tsPast, err := strconv.ParseInt(v.Timestamp, 10, 64)
		if err != nil {
			p.log.Error("handleGetAuthentication timestamp fmt error")
			return errRes, errors.New("handleGetAuthentication timestamp fmt error")
		}

		if math.Abs(float64(tsNow-tsPast)) >= config.DefautltAuthTimeout {
			p.log.Error("handleGetAuthentication timestamp expired")
			return errRes, errors.New("handleGetAuthentication timestamp expired")
		}

		data := hash.UsingSha256([]byte(v.PeerID + v.Addr + v.Timestamp))
		if ok, _ := cryptoClient.VerifyECDSA(publicKey, v.Sign, data); !ok {
			p.log.Error("handleGetAuthentication verify sign error")
			return errRes, errors.New("handleGetAuthentication verify sign error")
		}

		addrs = append(addrs, v.Addr)
	}

	resBuf, err := json.Marshal(addrs)
	if err != nil {
		p.log.Error("handleGetAuthentication json marshal error")
		return errRes, errors.New("handleGetAuthentication json marshal error")
	}

	p.log.Trace("handleGetAuthentication success", "logid", logid, "addrs", addrs)

	s.setReceivedAddr(addrs)
	s.isAuth = true

	res, err := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, "", logid,
		amper_p2p.AmperMessage_GET_AUTHENTICATION_RES, resBuf, amper_p2p.AmperMessage_SUCCESS)
	return res, err
}

func errorHandleGetAuthenMsg(logid string) *amper_p2p.AmperMessage {
	res, _ := amper_p2p.NewAmperMessage(amper_p2p.AmperMsgVersion2, "", logid,
		amper_p2p.AmperMessage_GET_AUTHENTICATION_RES, nil, amper_p2p.AmperMessage_GET_AUTHENTICATION_ERROR)
	return res
}
