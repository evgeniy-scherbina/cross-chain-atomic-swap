package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

func main() {
	btcd := launchBtcd("ShZTsTAgSQkmqZZHnU2mDKVCXP6h26Sm46")
	btcwallet := launchBtcwallet()
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
	time.Sleep(time.Second * 2)

	addr := btcctl()
	addr = strings.TrimSpace(addr)
	fmt.Printf("@%v@", addr)

	_ = btcd.Process.Kill()
	time.Sleep(time.Second * 2)
	cmd := launchBtcd(addr)
	_ = cmd
	time.Sleep(time.Second * 5)
	//fmt.Println(cmd.Wait())

	fmt.Println("HERE")

	if _, err := client.Generate(400); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 8)

	fmt.Println("HERE")

	createHtlc := flag.Bool("create_htlc", false, "")
	flag.Parse()
	if *createHtlc {
		privKey, previousTxHash, previousTx := receiveMoney(client)
		txHash, successPrivKey := sendHTLC(client, privKey, previousTxHash, previousTx)

		rPreImage := []byte{0x61,0x62,0x63,0x64}
		htlcSuccess(client, txHash, rPreImage, successPrivKey, successPrivKey.PubKey().SerializeCompressed())
	}

	_ = btcd.Process.Kill()
	_ = btcwallet.Process.Kill()
}
