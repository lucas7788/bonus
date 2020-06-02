package config

import (
	"github.com/ontio/bonus/utils"
	"github.com/ontio/ontology/common/log"
	"math/big"
)

var Version = ""

var (
	EthOneGwei            = new(big.Int).SetUint64(uint64(1000000000))
	ETH_DEFAULT_GAS_PRICE = utils.ToIntByPrecise("0.00000004", ETH_DECIMALS).Uint64() // 40 Gwei
	ETH_MIN_ETH_BANALNCE  = utils.ToIntByPrecise("0.00001", ETH_DECIMALS).Uint64()
	ONT_DEFAULT_GAS_LIMIT = 2000000
)

var DefConfig = &Config{
	RestPort:    DEFAULT_REST_PORT,
	Version:     "1.0.0",
	LevelDBPath: "./db/leveldb",
	OntCfg: &Ont{
		OntJsonRpcAddressMainNet: "http://dappnode1.ont.io:20336",
		OntJsonRpcAddressTestNet: "http://polaris1.ont.io:20336",
		GasPrice:                 500,
		GasLimit:                 2000000,
	},
	EthCfg: &Eth{
		RpcAddrMainNet: "http://onto-eth.ont.io:10331",
		//RpcAddrTestNet: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
		RpcAddrTestNet: "http://18.140.187.37:10331",
		GasPrice:       ETH_DEFAULT_GAS_PRICE,
	},
}

type Config struct {
	LevelDBPath string
	RestPort    uint   `json:"rest_port"`
	Version     string `json:"version"`
	OntCfg      *Ont   `json:"ont_cfg"`
	EthCfg      *Eth   `json:"eth_cfg"`
}

const (
	LogPath           = "./Log"
	DefaultWalletPath = "./wallet"
	DBPath            = "db"
)

var (
	DEFAULT_LOG_LEVEL = log.InfoLog
	DEFAULT_REST_PORT = uint(20327)

	ONT_TRANSFER_QUEUE_SIZE = 5000
	ETH_TRANSFER_QUEUE_SIZE = 500
	All_TOKEN_TYPE          []string //need init when server start, //TODO
)

type Ont struct {
	OntJsonRpcAddressTestNet string `json:"rpc_addr_test_net"`
	OntJsonRpcAddressMainNet string `json:"rpc_addr_main_net"`
	GasPrice                 uint64 `json:"gas_price"`
	GasLimit                 uint64 `json:"gas_limit"`
}

type Eth struct {
	RpcAddrTestNet string `json:"rpc_addr_test_net"`
	RpcAddrMainNet string `json:"rpc_addr_main_net"`
	GasPrice       uint64 `json:"gas_price"`
	GasLimit       uint64 `json:"gas_limit"`
}
type EthToken struct {
	TokenName    string `json:"token_name"`
	ContractAddr string `json:"contract_addr"`
}
