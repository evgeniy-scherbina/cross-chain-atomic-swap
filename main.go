package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
)

func main() {
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

	createHtlc := flag.Bool("create_htlc", false, "")
	flag.Parse()
	if *createHtlc {
		privKey, previousTxHash, previousTx := receiveMoney(client)
		txHash, successPrivKey := sendHTLC(client, privKey, previousTxHash, previousTx)

		rPreImage := []byte{0x61,0x62,0x63,0x64}
		//fmt.Println(len(rPreImage))
		//rawBytes := sha256.Sum256(rPreImage)
		//fmt.Println(hex.EncodeToString(rawBytes[:]))
		//return
		htlcSuccess(client, txHash, rPreImage, successPrivKey, successPrivKey.PubKey().SerializeCompressed())
	}
}
