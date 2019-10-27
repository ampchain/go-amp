/*
 * 
 */

package server

import (
	"github.com/ampchain/go-amp/core"
	"github.com/ampchain/go-amp/pb"
)

// HandleBlockCoreError core error <=> pb.error
func HandleBlockCoreError(err error) pb.AChainErrorEnum {
	switch err {
	case AmpChaincore.ErrCannotSyncBlock:
		return pb.AChainErrorEnum_CANNOT_SYNC_BLOCK_ERROR
	case AmpChaincore.ErrConfirmBlock:
		return pb.AChainErrorEnum_CONFIRM_BLOCK_ERROR
	case AmpChaincore.ErrUTXOVMPlay:
		return pb.AChainErrorEnum_UTXOVM_PLAY_ERROR
	case AmpChaincore.ErrWalk:
		return pb.AChainErrorEnum_WALK_ERROR
	case AmpChaincore.ErrNotReady:
		return pb.AChainErrorEnum_NOT_READY_ERROR
	case AmpChaincore.ErrBlockExist:
		return pb.AChainErrorEnum_BLOCK_EXIST_ERROR
	case AmpChaincore.ErrServiceRefused:
		return pb.AChainErrorEnum_SERVICE_REFUSED_ERROR
	default:
		return pb.AChainErrorEnum_UNKNOW_ERROR
	}
}
