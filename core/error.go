package AmpChaincore

import (
	"github.com/ampchain/go-amp/ledger"
	"github.com/ampchain/go-amp/pb"
	"github.com/ampchain/go-amp/utxo"
)

// HandlerUtxoError used to handle error of utxo
func HandlerUtxoError(err error) pb.AChainErrorEnum {
	switch err {
	case utxo.ErrAlreadyInUnconfirmed:
		return pb.AChainErrorEnum_UTXOVM_ALREADY_UNCONFIRM_ERROR
	case utxo.ErrNoEnoughUTXO:
		return pb.AChainErrorEnum_NOT_ENOUGH_UTXO_ERROR
	case utxo.ErrUTXONotFound:
		return pb.AChainErrorEnum_UTXOVM_NOT_FOUND_ERROR
	case utxo.ErrInputOutputNotEqual:
		return pb.AChainErrorEnum_INPUT_OUTPUT_NOT_EQUAL_ERROR
	case utxo.ErrTxNotFound:
		return pb.AChainErrorEnum_TX_NOT_FOUND_ERROR
	case utxo.ErrTxSizeLimitExceeded:
		return pb.AChainErrorEnum_TX_SLE_ERROR
	case utxo.ErrRWSetInvalid:
		return pb.AChainErrorEnum_RWSET_INVALID_ERROR
	default:
		return pb.AChainErrorEnum_UNKNOW_ERROR
	}
}

// HandlerLedgerError used to handle error of ledger
func HandlerLedgerError(err error) pb.AChainErrorEnum {
	switch err {
	case ledger.ErrRootBlockAlreadyExist:
		return pb.AChainErrorEnum_ROOT_BLOCK_EXIST_ERROR
	case ledger.ErrTxDuplicated:
		return pb.AChainErrorEnum_TX_DUPLICATE_ERROR
	case ledger.ErrTxNotFound:
		return pb.AChainErrorEnum_TX_NOT_FOUND_ERROR
	default:
		return pb.AChainErrorEnum_UNKNOW_ERROR
	}
}
