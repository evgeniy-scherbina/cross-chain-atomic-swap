package simnet

import (
	"fmt"
	"os"
	"os/exec"
)

type Btcd struct {
	cmd *exec.Cmd
}

func LaunchBtcd(miningAddr string) (*Btcd, error) {
	cmd := exec.Command(
		"btcd",
		"--simnet",
		fmt.Sprintf("--datadir=%v", btcdDataDir),
		fmt.Sprintf("--logdir=%v", btcdLogsDir),
		fmt.Sprintf("--listen=%v", btcdPeerListen),
		fmt.Sprintf("--rpclisten=%v", btcdRPCListen),
		fmt.Sprintf("--rpcuser=%v", btcdRPCUser),
		fmt.Sprintf("--rpcpass=%v", btcdRPCPass),
		fmt.Sprintf("--rpccert=%v", btcdRPCCert),
		fmt.Sprintf("--rpckey=%v", btcdRPCKey),
		fmt.Sprintf("--miningaddr=%v", miningAddr),
	)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &Btcd{
		cmd: cmd,
	}, nil
}

func (btcd *Btcd) Cmd() *exec.Cmd {
	return btcd.cmd
}

func (btcd *Btcd) Stop() {
	_ = btcd.Cmd().Process.Signal(os.Interrupt)
}

func (btcd *Btcd) Btcctl(args ...string) (string, error) {
	cmd := exec.Command(
		"btcctl",
		append([]string{
			"--simnet",
			fmt.Sprintf("--rpcserver=%v", btcwalletRPCListen),
			fmt.Sprintf("--rpcuser=%v", btcdRPCUser),
			fmt.Sprintf("--rpcpass=%v", btcdRPCPass),
			fmt.Sprintf("--rpccert=%v", btcdRPCCert),
			fmt.Sprintf("-C=%v", btcdDataDir),
		}, args...)...
	)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}