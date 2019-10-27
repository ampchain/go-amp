// +build wasm

package driver

import (
	"github.com/ampchain/go-amp/contractsdk/go/code"
	"github.com/ampchain/go-amp/contractsdk/go/driver/wasm"
)

// Serve run contract in wasm environment
func Serve(contract code.Contract) {
	driver := wasm.New()
	driver.Serve(contract)
}
