/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package eosio

import (
	"github.com/astaxie/beego/config"
	"github.com/eoscanada/eos-go"
	"path/filepath"
	"testing"

	"github.com/blocktree/openwallet/v2/log"
)

func testNewWalletManager() *WalletManager {
	wm := NewWalletManager(nil)
	//读取配置
	absFile := filepath.Join("conf", "EOS.ini")
	//log.Debug("absFile:", absFile)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return nil
	}
	wm.LoadAssetsConfig(c)
	wm.client.Debug = true
	wm.Api.Debug = true
	return wm
}

func TestWalletManager_GetInfo(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.Api.GetInfo()
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetAccount(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.Api.GetAccount("hrt3arlcl354")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetBlock(t *testing.T) {
	wm := testNewWalletManager()
	//3152
	r, err := wm.Api.GetBlockByNum(104761649)
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetTransaction(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.Api.GetTransaction("c7199c599098c3900fa6ff14813dca5becbaa8b4ca85fc4d8fdda94cd1005ab0")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)

}

func TestWalletManager_GetABI(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.Api.GetABI("eosio.token")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetCurrencyBalance(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.Api.GetCurrencyBalance("alice", "EOS", "eostesterxrv")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetActions(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.Api.GetActions(eos.GetActionsRequest{
		AccountName: "bob",
		Pos:         0,
		Offset:      100,
	})
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
	log.Infof("tx: %d", len(r.Actions))
}
