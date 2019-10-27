/*
 * 
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ampchain/go-amp/global"
	"github.com/ampchain/go-amp/pb"
)

// StatusCommand status cmd
type StatusCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewStatusCommand new status cmd
func NewStatusCommand(cli *Cli) *cobra.Command {
	s := new(StatusCommand)
	s.cli = cli
	s.cmd = &cobra.Command{
		Use:   "status",
		Short: "Operate a command to get status of current AmpChain server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return s.printAChainStatus(ctx)
		},
	}
	return s.cmd
}

func (s *StatusCommand) printAChainStatus(ctx context.Context) error {
	client := s.cli.AChainClient()
	req := &pb.CommonIn{
		Header: global.GHeader(),
	}
	reply, err := client.GetSystemStatus(ctx, req)
	if err != nil {
		return err
	}
	if reply.Header.Error != pb.AChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}
	status := FromSystemStatusPB(reply.GetSystemsStatus())
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

func init() {
	AddCommand(NewStatusCommand)
}
