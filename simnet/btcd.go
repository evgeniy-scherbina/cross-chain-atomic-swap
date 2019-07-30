package simnet

import (
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultTimeout               = time.Second * 5
	TimeoutForProcessShutdowning = time.Second * 2

	BlocksForSegWitActivation = 400
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

func (btcd *Btcd) Restart(miningAddr string) (*Btcd, error) {
	btcd.Stop()
	time.Sleep(TimeoutForProcessShutdowning)

	btcd, err := LaunchBtcd(miningAddr)
	if err != nil {
		return nil, err
	}
	time.Sleep(DefaultTimeout)
	return btcd, nil
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
		}, args...)...,
	)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (btcd *Btcd) ActivateSegWit(client *rpcclient.Client) error {
	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		return err
	}
	log.Printf("Block count: %d", blockCount)

	if blockCount < BlocksForSegWitActivation {
		if _, err := client.Generate(BlocksForSegWitActivation); err != nil {
			return err
		}
		time.Sleep(DefaultTimeout)
	}
	return nil
}

func (btcd *Btcd) GetNewAddress() (string, error) {
	addr, err := btcd.Btcctl("getnewaddress")
	if err != nil {
		return "", err
	}
	addr = strings.TrimSpace(addr)
	return addr, nil
}
