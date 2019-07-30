package main

import (
	"flag"
	"fmt"
	"github.com/evgeniy-scherbina/cross-chain-atomic-swap/simnet"
	"log"
	"time"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func launchSimnetAndWait() (*simnet.Btcd, *simnet.Btcwallet) {
	btcd, err := simnet.LaunchBtcd("ShZTsTAgSQkmqZZHnU2mDKVCXP6h26Sm46")
	checkErr(err)

	btcwallet, err := simnet.LaunchBtcwallet()
	checkErr(err)

	time.Sleep(simnet.DefaultTimeout)
	return btcd, btcwallet
}

func shutdownSimnetAndWait(btcd *simnet.Btcd, btcwallet *simnet.Btcwallet) {
	btcd.Stop()
	btcwallet.Stop()

	time.Sleep(simnet.TimeoutForProcessShutdowning)
}

func main() {
	btcd, btcwallet := launchSimnetAndWait()

	//rawCert, err := ioutil.ReadFile("data/rpc.cert")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Connect to local bitcoin core RPC server using HTTP POST mode.
	//connCfg := &rpcclient.ConnConfig{
	//	Host:         "localhost:18554",
	//	User:         "devuser",
	//	Pass:         "devpass",
	//	HTTPPostMode: true,  // Bitcoin core only supports HTTP POST mode
	//	DisableTLS:   false, // Bitcoin core does not provide TLS by default
	//	Certificates: rawCert,
	//}
	//// Notice the notification parameter is nil since notifications are
	//// not supported in HTTP POST mode.
	//client, err := rpcclient.New(connCfg, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer client.Shutdown()

	if err := btcd.RPCClient().WalletPassphrase("11111111", 3600 * 10); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 8)

	miningAddr, err := btcd.GetNewAddress()
	checkErr(err)
	fmt.Printf("Mining address %v\n", miningAddr)

	btcd, err = btcd.Restart(miningAddr)
	checkErr(err)

	checkErr(btcd.ActivateSegWit(btcd.RPCClient()))

	createHtlc := flag.Bool("create_htlc", false, "")
	flag.Parse()
	if *createHtlc {
		privKey, previousTxHash, previousTx := receiveMoney(btcd.RPCClient())
		txHash, successPrivKey := sendHTLC(btcd.RPCClient(), privKey, previousTxHash, previousTx)

		rPreImage := []byte{0x61,0x62,0x63,0x64}
		htlcSuccess(btcd.RPCClient(), txHash, rPreImage, successPrivKey, successPrivKey.PubKey().SerializeCompressed())
	}

	shutdownSimnetAndWait(btcd, btcwallet)
}
