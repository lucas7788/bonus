package config

import (
	"github.com/ontio/ontology-crypto/keypair"
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
	ProjectDBHost      string `json:"projectdb_host"`
	ProjectDBPort      string `json:"projectdb_port"`
	ProjectDBUrl       string `json:"projectdb_url"`
	ProjectDBUser      string `json:"projectdb_user"`
	ProjectDBPassword  string `json:"projectdb_password"`
	ProjectDBName      string `json:"projectdb_name"`
	Ont                *Ont   `json:"ont"`
	EthCfg             *Eth   `json:"eth"`
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
	OperatorPubKey         keypair.PublicKey
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
