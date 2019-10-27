// +build !wasm

package driver

import (
	"github.com/ampchain/go-amp/contractsdk/go/code"
	"github.com/ampchain/go-amp/contractsdk/go/driver/native"
)

// Serve run contract in native environment
func Serve(contract code.Contract) {
	driver := native.New()
	driver.Serve(contract)
}
