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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/blocktree/openwallet/common"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/eoscanada/eos-go"
)

const (
	blockchainBucket = "blockchain" // blockchain dataset
	//periodOfTask      = 5 * time.Second // task interval
	maxExtractingSize = 10 // thread count
)

//EOSBlockScanner EOS block scanner
type EOSBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
}

//ExtractResult extract result
type ExtractResult struct {
	extractData map[string][]*openwallet.TxExtractData
	TxID        string
	BlockHash   string
	BlockHeight uint64
	BlockTime   int64
	Success     bool
}

//SaveResult result
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

// NewEOSBlockScanner create a block scanner
func NewEOSBlockScanner(wm *WalletManager) *EOSBlockScanner {
	bs := EOSBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = true
	bs.RescanLastBlockCount = 0

	// set task
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

// ScanBlockTask scan block task
func (bs *EOSBlockScanner) ScanBlockTask() {

	var (
		currentHeight uint32
		currentHash   string
	)

	// get local block header
	currentHeight, currentHash, err := bs.GetLocalBlockHead()

	if err != nil {
		bs.wm.Log.Std.Error("", err)
	}

	if currentHeight == 0 {
		bs.wm.Log.Std.Info("No records found in local, get current block as the local!")

		headBlock, err := bs.GetGlobalHeadBlock()
		if err != nil {
			bs.wm.Log.Std.Info("get head block error, err=%v", err)
		}

		currentHash = headBlock.Previous.String()
		currentHeight = headBlock.BlockNum - 1
	}

	for {
		if !bs.Scanning {
			// stop scan
			return
		}

		infoResp, err := bs.GetChainInfo()
		if err != nil {
			bs.wm.Log.Errorf("get chain info failed, err=%v", err)
			break
		}

		maxBlockHeight := infoResp.HeadBlockNum

		bs.wm.Log.Info("current block height:", currentHeight, " maxBlockHeight:", maxBlockHeight)
		if currentHeight == maxBlockHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height %d", maxBlockHeight)
			break
		}

		// next block
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)
		block, err := bs.wm.Api.GetBlockByNum(currentHeight)

		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data by rpc; unexpected error: %v", err)
			break
		}

		if currentHash != block.Previous.String() {
			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previous.String())
			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			// get local fork bolck
			forkBlock, _ := bs.GetLocalBlock(currentHeight - 1)
			// delete last unscan block
			bs.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 // scan back to last 2 block
			if currentHeight <= 0 {
				currentHeight = 1
			}
			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner can not get local block; unexpected error: %v", err)
				//get block from rpc
				bs.wm.Log.Info("block scanner prev block height:", currentHeight)
				curBlock, err := bs.wm.Api.GetBlockByNum(currentHeight)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block by rpc; unexpected error: %v", err)
					break
				}
				currentHash = curBlock.ID.String()
			} else {
				//重置当前区块的hash
				currentHash = localBlock.Hash
			}
			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalBlockHead(currentHeight, currentHash)

			if forkBlock != nil {
				//通知分叉区块给观测者，异步处理
				bs.forkBlockNotify(forkBlock)
			}

		} else {
			currentHash = block.ID.String()
			err := bs.BatchExtractTransactions(uint64(currentHeight), currentHash, block.Timestamp.Unix(), block.Transactions)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner ran BatchExtractTransactions occured unexpected error: %v", err)
			}

			//保存本地新高度
			bs.SaveLocalBlockHead(currentHeight, currentHash)
			bs.SaveLocalBlock(ParseBlock(block))
			//通知新区块给观测者，异步处理
			bs.newBlockNotify(block)
		}
	}
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *EOSBlockScanner) forkBlockNotify(block *Block) {
	header := block.BlockHeader
	header.Fork = true
	bs.NewBlockNotify(&header)
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *EOSBlockScanner) newBlockNotify(block *eos.BlockResp) {
	header := ParseHeader(block)
	bs.NewBlockNotify(header)
}

// BatchExtractTransactions 批量提取交易单
func (bs *EOSBlockScanner) BatchExtractTransactions(blockHeight uint64, blockHash string, blockTime int64, transactions []eos.TransactionReceipt) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(transactions) //需要完成的总数
	)

	if len(transactions) == 0 {
		return nil
	}

	bs.wm.Log.Std.Info("block scanner ready extract transactions total: %d ", len(transactions))

	//生产通道
	producer := make(chan ExtractResult)
	defer close(producer)

	//消费通道
	worker := make(chan ExtractResult)
	defer close(worker)

	//保存工作
	saveWork := func(height uint64, result chan ExtractResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {
				notifyErr := bs.newExtractDataNotify(height, gets.extractData)
				if notifyErr != nil {
					failed++ //标记保存失败数
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "")
				bs.SaveUnscanRecord(unscanRecord)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eblockHeight uint64, eBlockHash string, eBlockTime int64, mTransactions []eos.TransactionReceipt, eProducer chan ExtractResult) {
		for _, tx := range mTransactions {
			bs.extractingCH <- struct{}{}
			go func(mBlockHeight uint64, mTx *eos.TransactionReceipt, end chan struct{}, mProducer chan<- ExtractResult) {
				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, eBlockTime, mTx, bs.ScanTargetFunc)
				//释放
				<-end

			}(eblockHeight, &tx, bs.extractingCH, eProducer)
		}
	}
	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(blockHeight, worker)

	//独立线程运行生产
	go extractWork(blockHeight, blockHash, blockTime, transactions, producer)

	//以下使用生产消费模式
	bs.extractRuntime(producer, worker, quit)

	if failed > 0 {
		return fmt.Errorf("block scanner saveWork failed")
	}

	return nil
}

//extractRuntime 提取运行时
func (bs *EOSBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

	var (
		values = make([]ExtractResult, 0)
	)

	for {
		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult
		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]
		}
		select {
		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			return
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}
	//return
}

// ExtractTransaction 提取交易单
func (bs *EOSBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, blockTime int64, transaction *eos.TransactionReceipt, scanTargetFunc openwallet.BlockScanTargetFunc) ExtractResult {
	var (
		success = true
		result  = ExtractResult{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			TxID:        transaction.Transaction.ID.String(),
			extractData: make(map[string][]*openwallet.TxExtractData),
			BlockTime:   blockTime,
		}
	)

	if transaction.Status != eos.TransactionStatusExecuted {
		bs.wm.Log.Std.Debug("transaction did not executed: %s", transaction.Transaction.ID)
		return ExtractResult{Success: true}
	}

	//提出交易单明细
	if transaction.Transaction.Packed == nil {
		bs.wm.Log.Std.Debug("trx packed empty: %s", transaction.Transaction.ID)
		return ExtractResult{Success: true}
	}
	signedTransaction, _ := transaction.Transaction.Packed.Unpack()

	for _, action := range signedTransaction.Actions {

		if action.Name == "transfer" {

			abiInfo, err := bs.wm.ContractDecoder.GetABIInfo(string(action.Account))
			if err != nil {
				bs.wm.Log.Std.Error("get ABI: %s", err)
				return ExtractResult{Success: false}
			}
			// use abi to get data bytes
			abi, ok := abiInfo.ABI.(*eos.ABI)
			if !ok {
				bs.wm.Log.Std.Error("convert abi error")
				return ExtractResult{Success: false}
			}
			bytes, _ := abi.DecodeAction(action.HexData, action.Name)

			var data TransferData

			_err := json.Unmarshal(bytes, &data)
			if _err != nil {
				bs.wm.Log.Std.Error("parse data error: %s", _err)
				return ExtractResult{Success: false}
			}

			if scanTargetFunc == nil {
				bs.wm.Log.Std.Error("scanTargetFunc is not configurated")
				return ExtractResult{Success: false}
			}

			//订阅地址为交易单中的发送者
			accountID1, ok1 := scanTargetFunc(openwallet.ScanTarget{Alias: data.From, Symbol: bs.wm.Symbol(), BalanceModelType: openwallet.BalanceModelTypeAccount})
			//订阅地址为交易单中的接收者
			accountID2, ok2 := scanTargetFunc(openwallet.ScanTarget{Alias: data.To, Symbol: bs.wm.Symbol(), BalanceModelType: openwallet.BalanceModelTypeAccount})
			if accountID1 == accountID2 && len(accountID1) > 0 && len(accountID2) > 0 {
				bs.InitExtractResult(accountID1, TransferAction{action, data}, &result, 0)
			} else {
				if ok1 {
					bs.InitExtractResult(accountID1, TransferAction{action, data}, &result, 1)
				}

				if ok2 {
					bs.InitExtractResult(accountID2, TransferAction{action, data}, &result, 2)
				}
			}

		}
	}
	result.Success = success
	return result

}

//InitExtractResult optType = 0: 输入输出提取，1: 输入提取，2：输出提取
func (bs *EOSBlockScanner) InitExtractResult(sourceKey string, action TransferAction, result *ExtractResult, optType int64) {

	data := action.TransferData

	txExtractDataArray := result.extractData[sourceKey]
	if txExtractDataArray == nil {
		txExtractDataArray = make([]*openwallet.TxExtractData, 0)
	}

	txExtractData := &openwallet.TxExtractData{}

	status := "1"
	reason := ""

	symbol := data.Quantity.Symbol.Symbol
	decimals := int32(data.Quantity.Symbol.Precision)
	amount := common.NewString(data.Quantity.Amount).String()

	contractID := openwallet.GenContractID(bs.wm.Symbol(), string(action.Account)+":"+symbol)
	coin := openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: true,
		ContractID: contractID,
	}

	coin.Contract = openwallet.SmartContract{
		Symbol:     bs.wm.Symbol(),
		ContractID: contractID,
		Address:    string(action.Account),
		Token:      symbol,
	}

	transx := &openwallet.Transaction{
		Fees:        "0",
		Coin:        coin,
		BlockHash:   result.BlockHash,
		BlockHeight: result.BlockHeight,
		TxID:        result.TxID,
		Decimal:     decimals,
		Amount:      amount,
		ConfirmTime: result.BlockTime,
		From:        []string{data.From + ":" + amount},
		To:          []string{data.To + ":" + amount},
		IsMemo:      true,
		Status:      status,
		Reason:      reason,
	}

	transx.SetExtParam("memo", data.Memo)

	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID

	txExtractData.Transaction = transx
	if optType == 0 {
		bs.extractTxInput(action, txExtractData)
		bs.extractTxOutput(action, txExtractData)
	} else if optType == 1 {
		bs.extractTxInput(action, txExtractData)
	} else if optType == 2 {
		bs.extractTxOutput(action, txExtractData)
	}

	txExtractDataArray = append(txExtractDataArray, txExtractData)
	result.extractData[sourceKey] = txExtractDataArray
}

//extractTxInput 提取交易单输入部分,无需手续费，所以只包含1个TxInput
func (bs *EOSBlockScanner) extractTxInput(action TransferAction, txExtractData *openwallet.TxExtractData) {

	tx := txExtractData.Transaction
	data := action.TransferData
	coin := openwallet.Coin(tx.Coin)

	//主网from交易转账信息，第一个TxInput
	txInput := &openwallet.TxInput{}
	txInput.Recharge.Sid = openwallet.GenTxInputSID(tx.TxID, bs.wm.Symbol(), coin.ContractID, uint64(0))
	txInput.Recharge.TxID = tx.TxID
	txInput.Recharge.Address = data.From
	txInput.Recharge.Coin = coin
	txInput.Recharge.Amount = tx.Amount
	txInput.Recharge.Symbol = coin.Symbol
	//txInput.Recharge.IsMemo = true
	//txInput.Recharge.Memo = data.Memo
	txInput.Recharge.BlockHash = tx.BlockHash
	txInput.Recharge.BlockHeight = tx.BlockHeight
	txInput.Recharge.Index = 0 //账户模型填0
	txInput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)
}

//extractTxOutput 提取交易单输入部分,只有一个TxOutPut
func (bs *EOSBlockScanner) extractTxOutput(action TransferAction, txExtractData *openwallet.TxExtractData) {

	tx := txExtractData.Transaction
	data := action.TransferData
	coin := openwallet.Coin(tx.Coin)

	//主网to交易转账信息,只有一个TxOutPut
	txOutput := &openwallet.TxOutPut{}
	txOutput.Recharge.Sid = openwallet.GenTxOutPutSID(tx.TxID, bs.wm.Symbol(), coin.ContractID, uint64(0))
	txOutput.Recharge.TxID = tx.TxID
	txOutput.Recharge.Address = data.To
	txOutput.Recharge.Coin = coin
	txOutput.Recharge.Amount = tx.Amount
	txOutput.Recharge.Symbol = coin.Symbol
	//txOutput.Recharge.IsMemo = true
	//txOutput.Recharge.Memo = data.Memo
	txOutput.Recharge.BlockHash = tx.BlockHash
	txOutput.Recharge.BlockHeight = tx.BlockHeight
	txOutput.Recharge.Index = 0 //账户模型填0
	txOutput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
}

//newExtractDataNotify 发送通知
func (bs *EOSBlockScanner) newExtractDataNotify(height uint64, extractData map[string][]*openwallet.TxExtractData) error {
	for o := range bs.Observers {
		for key, array := range extractData {
			for _, item := range array {
				err := o.BlockExtractDataNotify(key, item)
				if err != nil {
					log.Error("BlockExtractDataNotify unexpected error:", err)
					//记录未扫区块
					unscanRecord := NewUnscanRecord(height, "", "ExtractData Notify failed.")
					err = bs.SaveUnscanRecord(unscanRecord)
					if err != nil {
						log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
					}

				}
			}

		}
	}

	return nil
}

//ScanBlock 扫描指定高度区块
func (bs *EOSBlockScanner) ScanBlock(height uint64) error {

	block, err := bs.scanBlock(height)
	if err != nil {
		return err
	}

	//通知新区块给观测者，异步处理
	bs.newBlockNotify(block)

	return nil
}

func (bs *EOSBlockScanner) scanBlock(height uint64) (*eos.BlockResp, error) {

	block, err := bs.wm.Api.GetBlockByNum(uint32(height))
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		bs.wm.Log.Std.Info("block height: %d extract failed.", height)
		return nil, err
	}

	bs.wm.Log.Std.Info("block scanner scanning height: %d ...", block.ID.String())

	err = bs.BatchExtractTransactions(uint64(block.BlockNum), block.ID.String(), block.Timestamp.Unix(), block.Transactions)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	return block, nil
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *EOSBlockScanner) SetRescanBlockHeight(height uint64) error {
	if height <= 0 {
		return errors.New("block height to rescan must greater than 0. ")
	}

	block, err := bs.wm.Api.GetBlockByNum(uint32(height - 1))
	if err != nil {
		return err
	}

	bs.SaveLocalBlockHead(uint32(height-1), block.ID.String())

	return nil
}

// GetGlobalMaxBlockHeight GetGlobalMaxBlockHeight
func (bs *EOSBlockScanner) GetGlobalMaxBlockHeight() uint64 {
	headBlock, err := bs.GetGlobalHeadBlock()
	if err != nil {
		bs.wm.Log.Std.Info("get global head block error;unexpected error:%v", err)
		return 0
	}
	return uint64(headBlock.BlockNum)
}

//GetGlobalHeadBlock GetGlobalHeadBlock
func (bs *EOSBlockScanner) GetGlobalHeadBlock() (block *eos.BlockResp, err error) {
	infoResp, err := bs.GetChainInfo()
	if err != nil {
		bs.wm.Log.Std.Info("get chain info error;unexpected error:%v", err)
		return
	}

	block, err = bs.wm.Api.GetBlockByNum(infoResp.HeadBlockNum - 1)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get block by height; unexpected error:%v", err)
		return
	}

	return
}

//GetChainInfo GetChainInfo
func (bs *EOSBlockScanner) GetChainInfo() (infoResp *eos.InfoResp, err error) {
	infoResp, err = bs.wm.Api.GetInfo()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get info; unexpected error:%v", err)
	}
	return
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *EOSBlockScanner) GetScannedBlockHeight() uint64 {
	height, _, _ := bs.GetLocalBlockHead()
	return uint64(height)
}

//GetBalanceByAddress 查询地址余额
func (bs *EOSBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

	addrBalanceArr := make([]*openwallet.Balance, 0)

	return addrBalanceArr, nil
}

func (bs *EOSBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {
	infoResp, err := bs.GetChainInfo()
	if err != nil {
		bs.wm.Log.Std.Info("get chain info error;unexpected error:%v", err)
		return nil, err
	}
	return &openwallet.BlockHeader{Height: uint64(infoResp.HeadBlockNum), Hash: infoResp.HeadBlockID.String()}, nil
}
