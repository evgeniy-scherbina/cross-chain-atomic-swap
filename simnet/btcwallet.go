package simnet

import (
	"fmt"
	"os/exec"
)

type btcwallet struct {
	cmd *exec.Cmd
}

func LaunchBtcwallet() (*btcwallet, error) {
	cmd := exec.Command(
		"btcwallet",
		"--simnet",
		fmt.Sprintf("--appdata=%v", btcwalletDataDir),
		fmt.Sprintf("--rpcconnect=%v", btcwalletRPCConnect),
		fmt.Sprintf("--btcdusername=%v", btcdRPCUser),
		fmt.Sprintf("--btcdpassword=%v", btcdRPCPass),
		fmt.Sprintf("--rpclisten=%v", btcwalletRPCListen),
		fmt.Sprintf("--username=%v", btcdRPCUser),
		fmt.Sprintf("--password=%v", btcdRPCPass),
		fmt.Sprintf("--rpccert=%v", btcdRPCCert),
		fmt.Sprintf("--rpckey=%v", btcdRPCKey),
		fmt.Sprintf("--cafile=%v", btcdRPCCert),
	)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &btcwallet{
		cmd: cmd,
	}, nil
}

func (btcwallet *btcwallet) Cmd() *exec.Cmd {
	return btcwallet.cmd
}