package main

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

func showTx(msgTx *wire.MsgTx) string {
	tmpl := `
==============================================================================================================
	Version:  %v
	TxIn:     %v
	TxOut:    %v
	LockTime: %v
==============================================================================================================
	`
	return fmt.Sprintf(
		tmpl,
		msgTx.Version,
		showTxIns(msgTx.TxIn),
		showTxOuts(msgTx.TxOut),
		msgTx.LockTime,
	)
}

func showTxIns(txIns []*wire.TxIn) string {
	rez := ""
	for _, txIn := range txIns {
		rez += showTxIn(txIn)
	}
	return rez
}

func showTxIn(txIn *wire.TxIn) string {
	tmpl := `
	PreviousOutPoint: %v
	SignatureScript:  %v
	Witness:          %v
	Sequence:         %v
	`
	return fmt.Sprintf(
		tmpl,
		txIn.PreviousOutPoint.String(),
		hex.EncodeToString(txIn.SignatureScript),
		txIn.Witness,
		txIn.Sequence,
	)
}

func showTxOuts(txOuts []*wire.TxOut) string {
	rez := ""
	for _, txOut := range txOuts {
		rez += showTxOut(txOut)
	}
	return rez
}

func showTxOut(txOut *wire.TxOut) string {
	tmpl := `
	Value:    %v
	PkScript: %v
	`
	return fmt.Sprintf(tmpl, txOut.Value, txOut.PkScript)
}