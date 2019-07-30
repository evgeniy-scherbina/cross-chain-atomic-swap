package main

import (
	"fmt"
	"log"
	"os/exec"
)

func launchBtcd(miningAddr string) *exec.Cmd {
	cmd := exec.Command(
		"btcd",
		"--simnet",
		"--datadir=data",
		"--logdir=logs",
		"--listen=127.0.0.1:18444",
		"--rpclisten=127.0.0.1:18556",
		"--rpcuser=devuser",
		"--rpcpass=devpass",
		"--rpccert=data/rpc.cert",
		"--rpckey=data/rpc.key",
		fmt.Sprintf("--miningaddr=%v", miningAddr),
	)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	return cmd
}

func launchBtcwallet() *exec.Cmd {
	cmd := exec.Command(
		"btcwallet",
		"--simnet",
		"--appdata=btcwallet",
		"--rpcconnect=127.0.0.1:18556",
		"--btcdusername=devuser",
		"--btcdpassword=devpass",
		"--rpclisten=127.0.0.1:18554",
		"--username=devuser",
		"--password=devpass",
		"--rpccert=data/rpc.cert",
		"--rpckey=data/rpc.key",
		"--cafile=data/rpc.cert",
	)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	return cmd
}

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