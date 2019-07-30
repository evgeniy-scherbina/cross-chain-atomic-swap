package main

import (
	"log"
	"os/exec"
)

func btcctl() string {
	cmd := exec.Command(
		"btcctl",
		"--simnet",
		"--rpcserver=127.0.0.1:18554",
		"--rpcuser=devuser",
		"--rpcpass=devpass",
		"--rpccert=data/rpc.cert",
		"-C=data",
		"getnewaddress",
	)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(output)
}