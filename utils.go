package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"log"

	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

//OP_IF
//[HASHOP] <digest> OP_EQUALVERIFY OP_DUP OP_HASH160 <seller pubkey hash>
//	OP_ELSE
//<num> [TIMEOUTOP] OP_DROP OP_DUP OP_HASH160 <buyer pubkey hash>
//	OP_ENDIF
//OP_EQUALVERIFY
//OP_CHECKSIG
func makeHtlcScript(digest, timeoutPubkeyHash, successPubkeyHash []byte, timeout int64) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_IF)
	builder.AddOp(txscript.OP_SHA256)
	builder.AddData(digest)
	builder.AddOp(txscript.OP_EQUALVERIFY)
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(successPubkeyHash)
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(timeout)
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(timeoutPubkeyHash)
	builder.AddOp(txscript.OP_ENDIF)
	builder.AddOp(txscript.OP_EQUALVERIFY)
	builder.AddOp(txscript.OP_CHECKSIG)

	fmt.Println(showHtlcScript(digest, timeoutPubkeyHash, successPubkeyHash, timeout))
	return builder.Script()
}

func showHtlcScript(digest, timeoutPubkeyHash, successPubkeyHash []byte, timeout int64) string {
	name := "LOCKING_HTLC_SCRIPT"
	tmpl := `
%v
%v: BEGIN
	OP_IF
	OP_HASH256
	DIGEST: %v
	OP_EQUALVERIFY
	OP_DUP
	OP_HASH160
	SUCCESS_PUBKEY_HASH: %v
	OP_ELSE
	TIMEOUT: %v
	OP_CHECKLOCKTIMEVERIFY
	OP_DROP
	OP_DUP
	OP_HASH160
	TIMEOUT_PUBKEY_HASH: %v
	OP_ENDIF
	OP_EQUALVERIFY
	OP_CHECKSIG
%v: END
%v
	`
	return fmt.Sprintf(
		tmpl,
		defaultLineSeparator,
		name,
		hex.EncodeToString(digest),
		hex.EncodeToString(successPubkeyHash),
		timeout,
		hex.EncodeToString(timeoutPubkeyHash),
		name,
		defaultLineSeparator,
	)
}

func makeUnlockingHtlcScript(htlcSuccessInputScript, successPubKeyCompressed, rPreImage []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(htlcSuccessInputScript)
	builder.AddData(successPubKeyCompressed)
	builder.AddData(rPreImage)
	builder.AddOp(txscript.OP_TRUE)
	fmt.Println(showUnlockingHtlcScript(htlcSuccessInputScript, successPubKeyCompressed, rPreImage))
	return builder.Script()
}

func showUnlockingHtlcScript(htlcSuccessInputScript, successPubKeyCompressed, rPreImage []byte) string {
	name := "UNLOCKING_HTLC_SCRIPT"
	tmpl := `
%v
%v: BEGIN
	HTLC_SUCCESS_INPUT_SCRIPT: %v
	SUCCESS_PUB_KEY_COMPRESSED: %v
	R_PRE_IMAGE: %v
	OP_TRUE
%v: END
%v
	`
	return fmt.Sprintf(
		tmpl,
		defaultLineSeparator,
		name,
		hex.EncodeToString(htlcSuccessInputScript),
		hex.EncodeToString(successPubKeyCompressed),
		hex.EncodeToString(rPreImage),
		name,
		defaultLineSeparator,
	)
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}

func receiveMoney(client *rpcclient.Client) (*btcec.PrivateKey, *chainhash.Hash, *btcutil.Tx) {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Fatal(err)
	}
	pubKeyCompressed := privKey.PubKey().SerializeCompressed()
	pubKeyHash := hash160(pubKeyCompressed)

	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.SimNetParams)
	if err != nil {
		log.Fatal(err)
	}

	amt, err := btcutil.NewAmount(1)
	if err != nil {
		log.Fatal(err)
	}
	previousTxHash, err := client.SendToAddress(addr, amt)
	if err != nil {
		log.Fatal(err)
	}
	previousTx, err := client.GetRawTransaction(previousTxHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(showTx(previousTx.MsgTx(), "RECEIVE_MONEY_TRANSACTION"))

	if _, err := client.Generate(1); err != nil {
		log.Fatal(err)
	}

	return privKey, previousTxHash, previousTx
}

func sendHTLC(
	client *rpcclient.Client,
	privKey *btcec.PrivateKey,
	previousTxHash *chainhash.Hash,
	previousTx *btcutil.Tx,
) (*chainhash.Hash, *btcec.PrivateKey) {
	foundIndex := 0
	for index, output := range previousTx.MsgTx().TxOut {
		if output.Value == 100000000 {
			foundIndex = index
		}
	}
	fmt.Printf("DEBUG(FOUND_INDEX): %v\n", foundIndex)

	digest, err := hex.DecodeString("88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589")
	if err != nil {
		log.Fatal(err)
	}

	timeoutPrivKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Fatal(err)
	}
	timeoutPubkeyCompressed := timeoutPrivKey.PubKey().SerializeCompressed()
	timeoutPubkeyHash := hash160(timeoutPubkeyCompressed)

	successPrivKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Fatal(err)
	}
	successPubkeyCompressed := successPrivKey.PubKey().SerializeCompressed()
	successPubkeyHash := hash160(successPubkeyCompressed)

	htlcScript, err := makeHtlcScript(digest, timeoutPubkeyHash, successPubkeyHash, 0)
	if err != nil {
		log.Fatal(err)
	}
	msgTx := &wire.MsgTx{
		Version: 0,
		TxIn: []*wire.TxIn{
			&wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					Hash:  *previousTxHash,
					Index: uint32(foundIndex),
				},
				Sequence: 0,
			},
		},
		TxOut: []*wire.TxOut{
			&wire.TxOut{
				Value:    10000000,
				PkScript: htlcScript,
			},
		},
		LockTime: 0,
	}

	var buff bytes.Buffer
	if err := msgTx.Serialize(&buff); err != nil {
		log.Fatal(err)
	}

	signature, err := txscript.SignatureScript(msgTx, 0, previousTx.MsgTx().TxOut[foundIndex].PkScript, txscript.SigHashAll, privKey, true)
	if err != nil {
		log.Fatal(err)
	}
	msgTx.TxIn[0].SignatureScript = signature

	fmt.Println(showTx(msgTx, "HTLC_TRANSACTION"))

	txHash, err := client.SendRawTransaction(msgTx, true)
	if err != nil {
		log.Fatal(err)
	}
	return txHash, successPrivKey
}

func htlcSuccess(
	client *rpcclient.Client,
	txHash *chainhash.Hash,
	rPreImage []byte,
	successPrivKey *btcec.PrivateKey,
	successPubKeyCompressed []byte,
) {
	htlcTx, err := client.GetRawTransaction(txHash)
	if err != nil {
		log.Fatal(err)
	}

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	script, err := builder.Script()
	if err != nil {
		log.Fatal(err)
	}

	successHtlcTx := wire.MsgTx{
		Version: 0,
		TxIn: []*wire.TxIn{
			&wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					Hash:  *txHash,
					Index: 0,
				},
				Sequence: 0,
			},
		},
		TxOut: []*wire.TxOut{
			&wire.TxOut{
				Value:    10000000 - 10000,
				PkScript: script,
			},
		},
		LockTime: 0,
	}

	htlcSuccessInputScript, err := txscript.RawTxInSignature(
		&successHtlcTx,
		0,
		htlcTx.MsgTx().TxOut[0].PkScript,
		txscript.SigHashAll,
		successPrivKey,
	)

	script, err = makeUnlockingHtlcScript(htlcSuccessInputScript, successPubKeyCompressed, rPreImage)
	if err != nil {
		log.Fatal(err)
	}
	successHtlcTx.TxIn[0].SignatureScript = script

	fmt.Println(showTx(&successHtlcTx, "SUCCESS_HTLC_TRANSACTION"))

	execute(htlcTx.MsgTx().TxOut[0].PkScript, &successHtlcTx, 0, 10000000)

	if _, err := client.SendRawTransaction(&successHtlcTx, true); err != nil {
		log.Fatal(err)
	}
}
