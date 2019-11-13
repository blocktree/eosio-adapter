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
	"github.com/blocktree/openwallet/openwallet"
)

//SaveLocalBlockHead 记录区块高度和hash到本地
func (bs *EOSBlockScanner) SaveLocalBlockHead(blockHeight uint32, blockHash string) error {

	//获取本地区块高度
	//db, err := storm.Open(filepath.Join(bs.wm.Config.DBPath, bs.wm.Config.BlockchainFile))
	//if err != nil {
	//	return err
	//}
	//defer db.Close()
	//
	//db.Set(blockchainBucket, "blockHeight", &blockHeight)
	//db.Set(blockchainBucket, "blockHash", &blockHash)

	//return nil

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	header := &openwallet.BlockHeader{
		Hash:   blockHash,
		Height: uint64(blockHeight),
		Fork:   false,
		Symbol: bs.wm.Symbol(),
	}

	return bs.BlockchainDAI.SaveCurrentBlockHead(header)
}

//GetLocalBlockHead 获取本地记录的区块高度和hash
func (bs *EOSBlockScanner) GetLocalBlockHead() (uint32, string, error) {

	//var (
	//	blockHeight uint32
	//	blockHash   string
	//)

	//获取本地区块高度
	//db, err := storm.Open(filepath.Join(bs.wm.Config.DBPath, bs.wm.Config.BlockchainFile))
	//if err != nil {
	//	return 0, "", err
	//}
	//defer db.Close()
	//
	//db.Get(blockchainBucket, "blockHeight", &blockHeight)
	//db.Get(blockchainBucket, "blockHash", &blockHash)

	//return blockHeight, blockHash, nil

	if bs.BlockchainDAI == nil {
		return 0, "", fmt.Errorf("Blockchain DAI is not setup ")
	}

	header, err := bs.BlockchainDAI.GetCurrentBlockHead(bs.wm.Symbol())
	if err != nil {
		return 0, "", err
	}

	return uint32(header.Height), header.Hash, nil
}

//SaveLocalBlock 记录本地新区块
func (bs *EOSBlockScanner) SaveLocalBlock(blockHeader *Block) error {

	//db, err := storm.Open(filepath.Join(bs.wm.Config.DBPath, bs.wm.Config.BlockchainFile))
	//if err != nil {
	//	return err
	//}
	//defer db.Close()
	//
	//db.Save(blockHeader)
	//
	//return nil

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	header := &openwallet.BlockHeader{
		Hash:              blockHeader.Hash,
		Confirmations:     blockHeader.Confirmations,
		Merkleroot:        blockHeader.Merkleroot,
		Previousblockhash: blockHeader.Previousblockhash,
		Height:            uint64(blockHeader.Height),
		Version:           blockHeader.Version,
		Time:              blockHeader.Time,
		Fork:              blockHeader.Fork,
		Symbol:            bs.wm.Symbol(),
	}

	return bs.BlockchainDAI.SaveLocalBlockHead(header)
}

//GetLocalBlock 获取本地区块数据
func (bs *EOSBlockScanner) GetLocalBlock(height uint32) (*Block, error) {

	//var (
	//	blockHeader Block
	//)
	//
	//db, err := storm.Open(filepath.Join(bs.wm.Config.DBPath, bs.wm.Config.BlockchainFile))
	//if err != nil {
	//	return nil, err
	//}
	//defer db.Close()
	//
	//err = db.One("Height", height, &blockHeader)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return &blockHeader, nil

	if bs.BlockchainDAI == nil {
		return nil, fmt.Errorf("Blockchain DAI is not setup ")
	}

	header, err := bs.BlockchainDAI.GetLocalBlockHeadByHeight(uint64(height), bs.wm.Symbol())
	if err != nil {
		return nil, err
	}
	
	block := &Block{
		BlockHeader:  *header,
		Height:       uint32(header.Height),
		Fork:         header.Fork,
	}

	return block, nil
}

//SaveUnscanRecord 保存交易记录到钱包数据库
func (bs *EOSBlockScanner) SaveUnscanRecord(record *openwallet.UnscanRecord) error {

	//if record == nil {
	//	return errors.New("the unscan record to save is nil")
	//}
	//
	////获取本地区块高度
	//db, err := storm.Open(filepath.Join(bs.wm.Config.DBPath, bs.wm.Config.BlockchainFile))
	//if err != nil {
	//	return err
	//}
	//defer db.Close()
	//
	//return db.Save(record)

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.SaveUnscanRecord(record)
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (bs *EOSBlockScanner) DeleteUnscanRecord(height uint32) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.DeleteUnscanRecordByHeight(uint64(height), bs.wm.Symbol())
}

func (bs *EOSBlockScanner) GetUnscanRecords() ([]*openwallet.UnscanRecord, error) {

	if bs.BlockchainDAI == nil {
		return nil, fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.GetUnscanRecords(bs.wm.Symbol())
}