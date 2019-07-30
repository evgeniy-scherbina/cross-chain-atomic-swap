package main

import (
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/evgeniy-scherbina/cross-chain-atomic-swap/simnet"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	btcd, err := simnet.LaunchBtcd("ShZTsTAgSQkmqZZHnU2mDKVCXP6h26Sm46")
	checkErr(err)

	btcwallet, err := simnet.LaunchBtcwallet()
	checkErr(err)

	time.Sleep(time.Second * 5)

	rawCert, err := ioutil.ReadFile("data/rpc.cert")
	if err != nil {
		log.Fatal(err)
	}

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18554",
		User:         "devuser",
		Pass:         "devpass",
		HTTPPostMode: true,  // Bitcoin core only supports HTTP POST mode
		DisableTLS:   false, // Bitcoin core does not provide TLS by default
		Certificates: rawCert,
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)

	if err := client.WalletPassphrase("11111111", 3600 * 10); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 8)

	addr := btcctl()
	addr = strings.TrimSpace(addr)
	fmt.Printf("Mining address %v\n", addr)

	_ = btcd.Cmd().Process.Signal(os.Interrupt)
	time.Sleep(time.Second * 2)
	btcd, err = simnet.LaunchBtcd(addr)
	checkErr(err)
	time.Sleep(time.Second * 5)

	if blockCount < 400 {
		if _, err := client.Generate(400); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second * 8)
	}


	createHtlc := flag.Bool("create_htlc", false, "")
	flag.Parse()
	if *createHtlc {
		privKey, previousTxHash, previousTx := receiveMoney(client)
		txHash, successPrivKey := sendHTLC(client, privKey, previousTxHash, previousTx)

		rPreImage := []byte{0x61,0x62,0x63,0x64}
		htlcSuccess(client, txHash, rPreImage, successPrivKey, successPrivKey.PubKey().SerializeCompressed())
	}

	_ = btcd.Cmd().Process.Signal(os.Interrupt)
	_ = btcwallet.Cmd().Process.Signal(os.Interrupt)

	time.Sleep(time.Second * 2)
}
