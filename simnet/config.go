package simnet

import "fmt"

const (
	btcdDataDir = "data"
	btcdLogsDir = "logs"

	btcdPeerListen = "127.0.0.1:18444"
	btcdRPCListen  = "127.0.0.1:18556"

	btcdRPCUser = "devuser"
	btcdRPCPass = "devpass"


	btcwalletDataDir = "btcwallet"
	btcwalletRPCConnect = btcdRPCListen
)

var (
	btcdRPCCert = fmt.Sprintf("%v/rpc.cert", btcdDataDir)
	btcdRPCKey  = fmt.Sprintf("%v/rpc.key", btcdDataDir)

	btcwalletRPCListen  = "127.0.0.1:18554"
)