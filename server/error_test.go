package server

import (
	"errors"
	"github.com/ampchain/go-amp/core"
	"github.com/ampchain/go-amp/pb"
	"testing"
)

func TestHandleBlockCoreError(t *testing.T) {

	testCases := map[string]struct {
		in       error
		expected pb.AChainErrorEnum
	}{
		"1": {
			in:       AmpChaincore.ErrCannotSyncBlock,
			expected: pb.AChainErrorEnum_CANNOT_SYNC_BLOCK_ERROR,
		},
		"2": {
			in:       AmpChaincore.ErrConfirmBlock,
			expected: pb.AChainErrorEnum_CONFIRM_BLOCK_ERROR,
		},
		"3": {
			in:       AmpChaincore.ErrUTXOVMPlay,
			expected: pb.AChainErrorEnum_UTXOVM_PLAY_ERROR,
		},
		"4": {
			in:       AmpChaincore.ErrWalk,
			expected: pb.AChainErrorEnum_WALK_ERROR,
		},
		"5": {
			in:       AmpChaincore.ErrNotReady,
			expected: pb.AChainErrorEnum_NOT_READY_ERROR,
		},
		"6": {
			in:       AmpChaincore.ErrBlockExist,
			expected: pb.AChainErrorEnum_BLOCK_EXIST_ERROR,
		},
		"7": {
			in:       AmpChaincore.ErrServiceRefused,
			expected: pb.AChainErrorEnum_SERVICE_REFUSED_ERROR,
		},
		"8": {
			in:       errors.New("default"),
			expected: pb.AChainErrorEnum_UNKNOW_ERROR,
		},
	}
	for testName, testCase := range testCases {
		if actual := HandleBlockCoreError(testCase.in); testCase.expected != actual {
			t.Errorf("%s expected: %v, actual: %v", testName, testCase.expected, actual)
		}
	}
}
