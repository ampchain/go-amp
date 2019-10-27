package global

const (
	// SafeModel 表示安全的同步
	SafeModel = iota
	// Normal 表示正常状态
	Normal
)

const (
	// SRootChainName name of amper chain
	SRootChainName = "amper"
	// SBlockChainConfig configuration file name of amper chain
	SBlockChainConfig = "amper.json"
)

// XContext define the common context
type XContext struct {
	Timer *XTimer
}
