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
	"testing"
)

func TestSaveLocalBlockHead(t *testing.T) {
	height := uint32(50775550)
	hash := "0306c5fe30a97ad125c5edb3f8bf027f777f2f0b5cf34fba31974375f82ce08a"
	wm := testNewWalletManager()

	err := wm.Blockscanner.SaveLocalBlockHead(height, hash)
	if err != nil {
		t.Error(err)
	}
}

//GetLocalNewBlock 获取本地记录的区块高度和hash
func TestGetLocalNewBlock(t *testing.T) {
	height := uint32(50987586)
	hash := "030a0242b4fe359d247dd531a026d44dba62fdba7e38c765d86377ae7f68fc18"
	wm := testNewWalletManager()

	wm.Blockscanner.SaveLocalBlockHead(height, hash)

	_height, _hash, err := wm.Blockscanner.GetLocalBlockHead()

	if err != nil {
		t.Error(err)
		return
	}

	if _height != height || _hash != hash {
		t.Errorf("Expecting local height: %v and hash: %v, but returned height: %v and hash: %v.", height, hash, _height, _hash)
	}
}

//SaveLocalBlock 记录本地新区块
func TestSaveLocalBlock(t *testing.T) {
	wm := testNewWalletManager()
	block := Block{}

	block.Height = uint32(50990764)
	block.Hash = "030a0eac393c1899070c021ff5c4304e3790aefa89d0a6fe7e026dd86932f723"

	err := wm.Blockscanner.SaveLocalBlock(&block)

	if err != nil {
		t.Error(err)
	}
}

//GetLocalBlock 获取本地区块数据
func TestGetLocalBlock(t *testing.T) {
	height := 50991100
	hash := "030a0ffc276dc36dff0c35222d88fe3aecb1c02d99afe5f3705fff53cd2ebd93"
	wm := testNewWalletManager()
	block := Block{}

	block.Height = uint32(height)
	block.Hash = hash

	err := wm.Blockscanner.SaveLocalBlock(&block)
	if err != nil {
		t.Error(err)
		return
	}

	_block, err := wm.Blockscanner.GetLocalBlock(uint32(height))
	if err != nil {
		t.Error(err)
		return
	}

	if _block.Hash != block.Hash || _block.Height != block.Height {
		t.Errorf("Expecting block height: %v and hash: %v, but returned height: %v and hash: %v.", height, hash, _block.Height, _block.Hash)
	}
}
