package wasm

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ampchain/go-amp/common/config"
	"github.com/ampchain/go-amp/contract"
	"github.com/ampchain/go-amp/contract/bridge"
	"github.com/ampchain/go-amp/test/util"
)

type FakeWASMContext struct {
	*util.XModelContext
	vmm *VMManager
	vm  contract.VirtualMachine
}

func WithTestContext(t *testing.T, driver string, callback func(tctx *FakeWASMContext)) {
	util.WithXModelContext(t, func(x *util.XModelContext) {
		basedir := filepath.Join(x.Basedir, "wasm")
		xbridge := bridge.New()
		vmm, err := New(&config.WasmConfig{
			Driver: driver,
		}, basedir, xbridge, x.Model)
		if err != nil {
			t.Fatal(err)
		}
		exec := xbridge.RegisterExecutor("wasm", vmm)

		callback(&FakeWASMContext{
			vmm:           vmm,
			vm:            exec,
			XModelContext: x,
		})

	})
}

func loadWasmBinary(t *testing.T, filepath string) []byte {
	by, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}
	return by
}
