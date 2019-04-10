/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package eosio

import (
	"fmt"
	"testing"

	eos "github.com/eoscanada/eos-go"
)

//TestScanBlockTask
func TestScanBlockTask(t *testing.T) {
	wm := NewWalletManager()
	wm.Config.ServerAPI = "https://node1.zbeos.com"
	wm.Api = eos.New(wm.Config.ServerAPI)
	wm.Blockscanner.Scanning = true
	wm.Blockscanner.ScanBlockTask()
}

func TestEOSBlockScanner_ExtractTransaction(t *testing.T) {
	wm := NewWalletManager()
	wm.Config.ServerAPI = "https://node1.zbeos.com"
	wm.Api = eos.New(wm.Config.ServerAPI)
	wm.Blockscanner.Scanning = true

	bs := wm.Blockscanner

	blockHash := "031953833047e5085fdafc674077e019caa31c5e9dd6b50fbce224fee56d9084"
	blockResp, _ := wm.Api.GetBlockByID(blockHash)
	fmt.Println(blockResp.Transactions[7])

	result := bs.ExtractTransaction(uint64(blockResp.BlockNum), blockHash, blockResp.Timestamp.Unix(), &blockResp.Transactions[7], bs.ScanAddressFunc)

	fmt.Println(result)
}
