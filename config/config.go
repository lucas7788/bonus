package config

import (
	"github.com/ontio/ontology/common/log"
)

var Version = ""

var DefConfig = &Config{
	RestPort: DEFAULT_REST_PORT,
}

type Config struct {
	RestPort           uint   `json:"rest_port"`
	Version            string `json:"version"`
	HttpMaxConnections int    `json:"http_max_connections"`
	HttpCertPath       string `json:"http_cert_path"`
	HttpKeyPath        string `json:"http_key_path"`
	BonusDBHost        string `json:"bonusdb_host"`
	BonusDBPort        string `json:"bonusdb_port"`
	BonusDBUrl         string `json:"bonusdb_url"`
	BonusDBUser        string `json:"bonusdb_user"`
	BonusDBPassword    string `json:"bonusdb_password"`
	BonusDBName        string `json:"bonusdb_name"`
	OntCfg             *Ont   `json:"ont_cfg"`
	EthCfg             *Eth   `json:"eth_cfg"`
}

const (
	LogPath           = "./Log"
	DefaultWalletPath = "./wallet"
)

var (
	DEFAULT_LOG_LEVEL           = log.InfoLog
	DEFAULT_REST_PORT           = uint(20334)
	DEFAULT_HTTP_MAX_CONNECTION = 10000

	TRANSFER_QUEUE_SIZE    = 10000
	VERIFY_TX_QUEUE_SIZE   = 1000
	AIRDROP_QUEUE_SIZE     = 10000
	RECEIVE_INFO_CHAN_SIZE = 10000
	All_TOKEN_TYPE         []string //need init when server start, //TODO
)

type Ont struct {
	OntJsonRpcAddress string `json:"ont_json_rpc_address"`
	GasPrice          uint64 `json:"gas_price"`
	GasLimit          uint64 `json:"gas_limit"`
	WalletFile        string `json:"wallet_file"`
}

type Eth struct {
	KeyStore string `json:"key_store"`
	RpcAddr  string `json:"rpc_addr"`
}
type EthToken struct {
	TokenName    string `json:"token_name"`
	ContractAddr string `json:"contract_addr"`
}
