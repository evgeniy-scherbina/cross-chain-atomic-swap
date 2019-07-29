package main

import (
	"log"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func execute(scriptPubKey []byte, tx *wire.MsgTx, txIdx int, inputAmount int64) {

	engine, err := txscript.NewEngine(scriptPubKey, tx, txIdx, 0, nil, nil, inputAmount)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Execute(); err != nil {
		log.Fatal(engine)
	}
}