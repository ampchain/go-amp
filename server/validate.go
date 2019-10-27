/*
 * 
 */

package server

import (
	"errors"
	"github.com/ampchain/go-amp/pb"
)

func validateSendBlock(block *pb.Block) error {
	if len(block.Blockid) == 0 {
		return errors.New("validation error: validateSendBlock Block.Blockid can't be null")
	}

	if nil == block.Block {
		return errors.New("validation error: validateSendBlock Block.Block can't be null")
	}
	return nil
}
