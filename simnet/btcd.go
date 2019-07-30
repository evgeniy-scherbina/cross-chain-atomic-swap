package simnet

import (
	"fmt"
	"log"
	"os/exec"
)

func LaunchBtcd(miningAddr string) *exec.Cmd {
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
