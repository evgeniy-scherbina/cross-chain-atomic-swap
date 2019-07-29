package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"hash"
	"log"

	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"

	"github.com/btcsuite/btcd/txscript"
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

	fmt.Println("OP_IF")
	fmt.Println("OP_HASH256")
	fmt.Println("DIGEST", digest)
	fmt.Println("OP_EQUALVERIFY")
	fmt.Println("OP_DUP")
	fmt.Println("OP_HASH160")
	fmt.Println("SUCCESS_PUBKEY_HASH", successPubkeyHash)
	fmt.Println("OP_ELSE")
	fmt.Println("TIMEOUT", timeout)
	fmt.Println("OP_CHECKLOCKTIMEVERIFY")
	fmt.Println("OP_DROP")
	fmt.Println("OP_DUP")
	fmt.Println("OP_HASH160")
	fmt.Println("TIMEOUT_PUBKEY_HASH", timeoutPubkeyHash)
	fmt.Println("OP_ENDIF")
	fmt.Println("OP_EQUALVERIFY")
	fmt.Println("OP_CHECKSIG")
	return builder.Script()
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
	fmt.Println("RECEIVE_MONEY_TRANSACTION: BEGIN")
	fmt.Println(showTx(previousTx.MsgTx()))
	fmt.Println("RECEIVE_MONEY_TRANSACTION: END")

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
	fmt.Println("FOUND_INDEX", foundIndex)

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

	fmt.Println("HTLC_TRANSACTION: BEGIN")
	fmt.Println(showTx(msgTx))
	fmt.Println("HTLC_TRANSACTION: END")

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

	//foundIndex := 0
	//for index, output := range htlcTx.MsgTx().TxOut {
	//	if output.Value == 10000000 {
	//		foundIndex = index
	//	}
	//}
	//fmt.Println("FOUND_INDEX", foundIndex)

	//timeoutPrivKey, err := btcec.NewPrivateKey(btcec.S256())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//timeoutPubkeyCompressed := timeoutPrivKey.PubKey().SerializeCompressed()
	//timeoutPubkeyHash := hash160(timeoutPubkeyCompressed)

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

	builder = txscript.NewScriptBuilder()
	builder.AddData(htlcSuccessInputScript)
	builder.AddData(successPubKeyCompressed)
	builder.AddData(rPreImage)
	builder.AddOp(txscript.OP_TRUE)
	fmt.Println("HTLC_SUCCESS_INPUT_SCRIPT", htlcSuccessInputScript)
	fmt.Println("SUCCESS_PUB_KEY_COMPRESSED", successPubKeyCompressed)
	fmt.Println("R_PRE_IMAGE", rPreImage)
	fmt.Println("OP_TRUE")
	script, err = builder.Script()
	if err != nil {
		log.Fatal(err)
	}
	successHtlcTx.TxIn[0].SignatureScript = script

	fmt.Println("SUCCESS_HTLC_TRANSACTION: BEGIN")
	fmt.Println(showTx(&successHtlcTx))
	fmt.Println("SUCCESS_HTLC_TRANSACTION: END")

	//scriptPubKey []byte, tx *wire.MsgTx, txIdx int, inputAmount int64

	fmt.Println("EXECUTE: BEGIN")
	execute(htlcTx.MsgTx().TxOut[0].PkScript, &successHtlcTx, 0, 10000000)
	fmt.Println("EXECUTE: END")

	if _, err := client.SendRawTransaction(&successHtlcTx, true); err != nil {
		log.Fatal(err)
	}
}

//OP_IF
//[HASHOP] <digest> OP_EQUALVERIFY OP_DUP OP_HASH160 <seller pubkey hash>
//	OP_ELSE
//<num> [TIMEOUTOP] OP_DROP OP_DUP OP_HASH160 <buyer pubkey hash>
//	OP_ENDIF
//OP_EQUALVERIFY
//OP_CHECKSIG










//HTLC_SUCCESS_INPUT_SCRIPT [48 68 2 32 49 213 157 28 13 11 168 74 175 23 214 177 211 126 156 239 97 183 88 147 201 19 119 184 125 162 112 196 138 179 142 61 2 32 104 22 84 232 76 57 215 207 1 237 217 169 72 249 138 252 7 246 3 132 48 219 103 76 114 15 65 56 60 71 26 204 1]
//SUCCESS_PUB_KEY_COMPRESSED [2 109 5 239 223 102 229 6 227 62 252 13 115 174 33 2 166 191 173 93 228 100 223 170 197 173 221 116 168 247 157 70 137]
//R_PRE_IMAGE [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
//OP_TRUE


//OP_IF
//OP_HASH256
//DIGEST [40 137 173 170 213 129 250 232 138 106 26 231 243 51 93 222 96 191 234 242 150 58 26 61 193 193 197 170 45 234 236 133]
//OP_EQUALVERIFY
//OP_DUP
//OP_HASH160
//SUCCESS_PUBKEY_HASH [155 219 250 138 14 247 170 202 8 218 16 175 185 238 184 224 51 85 224 31]
//OP_ELSE
//TIMEOUT 0
//OP_CHECKLOCKTIMEVERIFY
//OP_DROP
//OP_DUP
//OP_HASH160
//TIMEOUT_PUBKEY_HASH [107 53 105 169 230 115 45 41 99 159 197 121 197 204 170 117 107 183 26 166]
//OP_ENDIF
//OP_EQUALVERIFY
//OP_CHECKSIG