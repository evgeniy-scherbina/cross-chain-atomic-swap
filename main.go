package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/evgeniy-scherbina/cross-chain-atomic-swap/simnet"
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

	err := btcd.UnlockWallet()
	checkErr(err)

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
