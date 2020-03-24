package config

const (
	ONT_DECIMALS = 0
	ONG_DECIMALS = 9
	ETH_DECIMALS = 18
	PASSWORD     = "ABCDEFGH"
)

const (
	TestNet = "testnet"
	MainNet = "mainnet"
)

const (
	ONT   = "ONT"
	ONG   = "ONG"
	OEP4  = "OEP4"
	ETH   = "ETH"
	ERC20 = "ERC20"
)

const (
	RetryLimit    = 6
	SleepTime     = 3
	EthSendTxSlot = 20
	EthSleepTime  = 6
	PendingLimit  = 100
)

var (
	SupportedTokenTypes = []string{ONT, ONG, OEP4, ETH, ERC20}
)
