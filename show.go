package main

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

const defaultLineSeparator = "====================================================================================="

func showTx(msgTx *wire.MsgTx, humanReadableName string) string {
	tmpl := `
%v
%v_BEGIN
	Version:  %v
	TxIn:     %v
	TxOut:    %v
	LockTime: %v
%v_END
%v
	`
	return fmt.Sprintf(
		tmpl,
		defaultLineSeparator,
		humanReadableName,
		msgTx.Version,
		showTxIns(msgTx.TxIn),
		showTxOuts(msgTx.TxOut),
		msgTx.LockTime,
		humanReadableName,
		defaultLineSeparator,
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
	return fmt.Sprintf(tmpl, txOut.Value, hex.EncodeToString(txOut.PkScript))
}