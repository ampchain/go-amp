// wrapper for consensus-tdpos plugin
package main

import (
	"github.com/ampchain/go-amp/consensus/tdpos"
)

// GetInstance : implement plugin framework
func GetInstance() interface{} {
	tdposIns := tdpos.TDpos{}
	tdposIns.Init()
	return &tdposIns
}
