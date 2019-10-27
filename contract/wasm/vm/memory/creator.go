package memory

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/ampchain/go-amp/contract"
	"github.com/ampchain/go-amp/contract/bridge"
	"github.com/ampchain/go-amp/contract/bridge/memrpc"
	"github.com/ampchain/go-amp/contract/wasm/vm"
	"github.com/ampchain/go-amp/contractsdk/go/code"
	"github.com/ampchain/go-amp/contractsdk/go/exec"
)

type memoryInstanceCreator struct {
	config    vm.InstanceCreatorConfig
	codeCache map[string]code.Contract
}

func newMemoryInstanceCreator(config *vm.InstanceCreatorConfig) (vm.InstanceCreator, error) {
	return &memoryInstanceCreator{
		config:    *config,
		codeCache: make(map[string]code.Contract),
	}, nil
}

func (m *memoryInstanceCreator) CreateInstance(ctx *bridge.Context, cp vm.ContractCodeProvider) (vm.Instance, error) {
	contract, ok := m.codeCache[ctx.ContractName]
	if !ok {
		codebuf, err := cp.GetContractCode(ctx.ContractName)
		if err != nil {
			return nil, err
		}
		contract, err = Decode(codebuf)
		if err != nil {
			return nil, err
		}
		m.codeCache[ctx.ContractName] = contract
	}
	return newMemoryInstance(contract, ctx, m.config.SyscallService), nil
}

func (m *memoryInstanceCreator) RemoveCache(contractName string) {
	delete(m.codeCache, contractName)
}

type memoryInstance struct {
	contract      code.Contract
	bridgeContext *bridge.Context
	rpcServer     *memrpc.Server
}

func newMemoryInstance(contract code.Contract, ctx *bridge.Context, syscall *bridge.SyscallService) *memoryInstance {
	return &memoryInstance{
		contract:      contract,
		bridgeContext: ctx,
		rpcServer:     memrpc.NewServer(syscall),
	}
}

func (m *memoryInstance) bridgeCall(method string, request proto.Message, response proto.Message) error {
	requestBuf, _ := proto.Marshal(request)
	responseBuf, err := m.rpcServer.CallMethod(context.TODO(), m.bridgeContext.ID, method, requestBuf)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(responseBuf, response)
	return err
}

func (m *memoryInstance) Exec(function string) error {
	exec.RunContract(m.bridgeContext.ID, m.contract, m.bridgeCall)
	return nil
}

func (m *memoryInstance) ResourceUsed() contract.Limits {
	return contract.Limits{}
}

func (m *memoryInstance) Release() {
}

func init() {
	vm.Register("memory", newMemoryInstanceCreator)
}
