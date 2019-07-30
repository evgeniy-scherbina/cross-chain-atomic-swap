package simnet

import (
	"fmt"
	"os"
	"os/exec"
)

type Btcwallet struct {
	cmd *exec.Cmd
}

func LaunchBtcwallet() (*Btcwallet, error) {
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
	return &Btcwallet{
		cmd: cmd,
	}, nil
}

func (btcwallet *Btcwallet) Cmd() *exec.Cmd {
	return btcwallet.cmd
}

func (btcwallet *Btcwallet) Stop() {
	_ = btcwallet.Cmd().Process.Signal(os.Interrupt)
}