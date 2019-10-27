package amperp2p

import (
	"hash/crc32"

	"github.com/ampchain/go-amp/global"
)

// define message versions
const (
	AmperMsgVersion1 = "1.0.0"
	AmperMsgVersion2 = "2.0.0"
)

// NewAmperMessage create P2P message instance with given params
func NewAmperMessage(version, bcName, lgid string, tp AmperMessage_MessageType, msgInfo []byte, ep AmperMessage_ErrorType) (*AmperMessage, error) {
	msg := &AmperMessage{
		Header: &AmperMessage_MessageHeader{
			Version: version,
			Bcname:  bcName,
			Type:    tp,
		},
		Data: &AmperMessage_MessageData{
			MsgInfo: msgInfo,
		},
	}
	if lgid == "" {
		msg.Header.Logid = global.Glogid()
	} else {
		msg.Header.Logid = lgid
	}
	if version > AmperMsgVersion1 {
		msg.Header.ErrorType = ep
	}
	msg.Header.DataCheckSum = CalDataCheckSum(msg)
	return msg, nil
}

// CalDataCheckSum calculate checksum of message
func CalDataCheckSum(msg *AmperMessage) uint32 {
	return crc32.ChecksumIEEE(msg.GetData().GetMsgInfo())
}

// VerifyDataCheckSum verify the checksum of message
func VerifyDataCheckSum(msg *AmperMessage) bool {
	return crc32.ChecksumIEEE(msg.GetData().GetMsgInfo()) == msg.GetHeader().GetDataCheckSum()
}

// VerifyMsgMatch 用于带返回的请求场景下验证收到的消息是否为预期的消息
func VerifyMsgMatch(msgRaw *AmperMessage, msgNew *AmperMessage, peerID string) bool {
	if msgNew.GetHeader().GetFrom() != peerID {
		return false
	}
	if msgRaw.GetHeader().GetLogid() != msgNew.GetHeader().GetLogid() {
		return false
	}
	switch msgRaw.GetHeader().GetType() {
	case AmperMessage_GET_BLOCK:
		if msgNew.GetHeader().GetType() == AmperMessage_GET_BLOCK_RES {
			return true
		}
		return false
	case AmperMessage_GET_BLOCKCHAINSTATUS:
		if msgNew.GetHeader().GetType() == AmperMessage_GET_BLOCKCHAINSTATUS_RES {
			return true
		}
		return false
	case AmperMessage_CONFIRM_BLOCKCHAINSTATUS:
		if msgNew.GetHeader().GetType() == AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES {
			return true
		}
		return false
	case AmperMessage_GET_AUTHENTICATION:
		if msgNew.GetHeader().GetType() == AmperMessage_GET_AUTHENTICATION_RES {
			return true
		}
		return false
	}

	return true
}

// GetResMsgType get the message type
func GetResMsgType(msgType AmperMessage_MessageType) AmperMessage_MessageType {
	switch msgType {
	case AmperMessage_GET_BLOCK:
		return AmperMessage_GET_BLOCK_RES
	case AmperMessage_GET_BLOCKCHAINSTATUS:
		return AmperMessage_GET_BLOCKCHAINSTATUS_RES
	case AmperMessage_CONFIRM_BLOCKCHAINSTATUS:
		return AmperMessage_CONFIRM_BLOCKCHAINSTATUS_RES
	case AmperMessage_GET_RPC_PORT:
		return AmperMessage_GET_RPC_PORT_RES
	case AmperMessage_GET_AUTHENTICATION:
		return AmperMessage_GET_AUTHENTICATION_RES
	default:
		return AmperMessage_MSG_TYPE_NONE
	}
}
