/*
 *
 */

package common

import (
	"errors"
	"strings"

	"github.com/ampchain/go-amp/pb"
)

var (
	// ErrContractExecutionTimeout common error for contract timeout
	ErrContractExecutionTimeout = errors.New("contract execution timeout")
	// ErrContractConnectionError connect error
	ErrContractConnectionError = errors.New("can't connect contract")
	ErrKVNotFound              = errors.New("Key not found")
)

// ServerError AmpChain.proto error
type ServerError struct {
	Errno pb.AChainErrorEnum
}

// Error convert to name
func (err ServerError) Error() string {
	return pb.AChainErrorEnum_name[int32(err.Errno)]
}

func NormalizedKVError(err error) error {
	if err == nil {
		return err
	}
	if strings.HasSuffix(err.Error(), "not found") {
		return ErrKVNotFound
	}
	return err
}
