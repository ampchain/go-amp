package vm

import (
	"github.com/ampchain/go-amp/common/log"
	"github.com/ampchain/go-amp/contract"
	"github.com/ampchain/go-amp/contract/bridge"
	"github.com/ampchain/go-amp/pb"
)

// InstanceCreatorConfig configures InstanceCreator
type InstanceCreatorConfig struct {
	Basedir        string
	SyscallService *bridge.SyscallService
	// VMConfig is the config of vm driver
	VMConfig    interface{}
	DebugLogger *log.Logger
}

// NewInstanceCreatorFunc instances a new InstanceCreator from InstanceCreatorConfig
type NewInstanceCreatorFunc func(config *InstanceCreatorConfig) (InstanceCreator, error)

// ContractCodeProvider provides source code and desc of contract
type ContractCodeProvider interface {
	GetContractCodeDesc(name string) (*pb.WasmCodeDesc, error)
	GetContractCode(name string) ([]byte, error)
}

// InstanceCreator is the creator of wasm virtual machine instance
type InstanceCreator interface {
	// CreateInstance instances a wasm virtual machine instance which can run a single contract call
	CreateInstance(ctx *bridge.Context, cp ContractCodeProvider) (Instance, error)
	RemoveCache(name string)
}

// Instance is a wasm virtual machine instance which can run a single contract call
type Instance interface {
	Exec(function string) error
	ResourceUsed() contract.Limits
	Release()
}
