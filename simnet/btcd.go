package simnet

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

const (
	DefaultTimeout               = time.Second * 5
	TimeoutForProcessShutdowning = time.Second * 2

	BlocksForSegWitActivation = 400
)

type Btcd struct {
	cmd *exec.Cmd
	rpcClient *rpcclient.Client
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
	btcd := &Btcd{
		cmd: cmd,
	}
	if err := btcd.enableRPCClient(); err != nil {
		return nil, err
	}
	return btcd, nil
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

func (btcd *Btcd) enableRPCClient() error {
	rawCert, err := ioutil.ReadFile("data/rpc.cert")
	if err != nil {
		return err
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
		return err
	}

	btcd.rpcClient = client
	return nil
}

func (btcd *Btcd) Cmd() *exec.Cmd {
	return btcd.cmd
}

func (btcd *Btcd) RPCClient() *rpcclient.Client {
	return btcd.rpcClient
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

func (btcd *Btcd) UnlockWallet() error {
	if err := btcd.RPCClient().WalletPassphrase("11111111", 3600 * 10); err != nil {
		return err
	}
	time.Sleep(DefaultTimeout)
	return nil
}