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
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/blocktree/eosio-adapter/eos_txsigner"
	"github.com/blocktree/go-owcrypt"

	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

// TransactionDecoder 交易单解析器
type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
	//TransferActionName string
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	//decoder.TransferActionName = "transfer"
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID      = rawTx.Account.AccountID
		accountBalance eos.Asset
		amountStr      string
		to             string
		codeAccount    string
		tokenCoin      string
	)

	addr := strings.Split(rawTx.Coin.Contract.Address, ":")
	if len(addr) != 2 {
		return fmt.Errorf("token contract's address is invalid: %s", rawTx.Coin.Contract.Address)
	}
	codeAccount = addr[0]
	tokenCoin = strings.ToUpper(addr[1])

	//获取wallet
	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return err
	}

	if account.Alias == "" {
		return fmt.Errorf("[%s] have not been created", accountID)
	}

	//账户是否上链
	accountResp, err := decoder.wm.Api.GetAccount(eos.AccountName(account.Alias))
	if err != nil && accountResp == nil {
		return fmt.Errorf("%s account of from not found on chain", decoder.wm.Symbol())
	}
	//fmt.Println("Permission for initn:", accountResp.Permissions[0].RequiredAuth.Keys[0])
	for k, v := range rawTx.To {
		amountStr = v
		to = k
		break
	}

	// 检查目标账户是否存在
	accountTo, err := decoder.wm.Api.GetAccount(eos.AccountName(to))
	if err != nil && accountTo == nil {
		return fmt.Errorf("%s account of to not found on chain", decoder.wm.Symbol())
	}

	accountAssets, err := decoder.wm.Api.GetCurrencyBalance(eos.AccountName(account.Alias), tokenCoin, eos.AccountName(codeAccount))
	if len(accountAssets) == 0 {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "all address's balance of account is not enough")
	}

	accountBalance = accountAssets[0]

	accountBalanceDec := decimal.New(int64(accountBalance.Amount), -int32(accountBalance.Precision))
	amountDec, _ := decimal.NewFromString(amountStr)

	if accountBalanceDec.LessThan(amountDec) {
		return fmt.Errorf("the balance: %s is not enough", amountStr)
	}

	amountInt64 := amountDec.Shift(int32(accountBalance.Precision)).IntPart()
	quantity := eos.Asset{Amount: eos.Int64(amountInt64), Symbol: accountBalance.Symbol}
	memo := rawTx.GetExtParam().Get("memo").String()

	createTxErr := decoder.createRawTransaction(
		wrapper,
		rawTx,
		accountResp,
		quantity,
		memo)
	if createTxErr != nil {
		return createTxErr
	}

	return nil

}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return err
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}
			hash, err := hex.DecodeString(keySignature.Message)
			if err != nil {
				return fmt.Errorf("decoder transaction hash failed, unexpected err: %v", err)
			}

			decoder.wm.Log.Debug("hash:", hash)

			sig, err := eos_txsigner.Default.SignTransactionHash(hash, keyBytes, decoder.wm.CurveType())
			if err != nil {
				return fmt.Errorf("sign transaction hash failed, unexpected err: %v", err)
			}

			keySignature.Signature = hex.EncodeToString(sig)
		}
	}

	decoder.wm.Log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	var tx eos.Transaction
	txHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}
	err = eos.UnmarshalBinary(txHex, &tx)
	if err != nil {
		return fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}

	stx := eos.NewSignedTransaction(&tx)

	//支持多重签名
	for accountID, keySignatures := range rawTx.Signatures {
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			messsage, _ := hex.DecodeString(keySignature.Message)
			signature, _ := hex.DecodeString(keySignature.Signature)
			publicKey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			//验证签名，解压公钥，解压后首字节04要去掉
			uncompessedPublicKey := owcrypt.PointDecompress(publicKey, decoder.wm.CurveType())

			valid, compactSig, err := eos_txsigner.Default.VerifyAndCombineSignature(messsage, uncompessedPublicKey[1:], signature)
			if !valid {
				return fmt.Errorf("transaction verify failed: %v", err)
			}

			stx.Signatures = append(
				stx.Signatures,
				ecc.Signature{Curve: ecc.CurveK1, Content: compactSig},
			)
		}
	}

	bin, err := eos.MarshalBinary(stx)
	if err != nil {
		return fmt.Errorf("signed transaction encode failed, unexpected error: %v", err)
	}

	rawTx.IsCompleted = true
	rawTx.RawHex = hex.EncodeToString(bin)

	return nil
}

// SubmitRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	var stx eos.SignedTransaction
	txHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return nil, fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}
	err = eos.UnmarshalBinary(txHex, &stx)
	if err != nil {
		return nil, fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}

	packedTx, err := stx.Pack(eos.CompressionNone)
	if err != nil {
		return nil, err
	}

	response, err := decoder.wm.BroadcastAPI.PushTransaction(packedTx)
	if err != nil {
		return nil, fmt.Errorf("push transaction: %s", err)
	}

	log.Infof("Transaction [%s] submitted to the network successfully.", hex.EncodeToString(response.Processed.ID))

	rawTx.TxID = hex.EncodeToString(response.Processed.ID)
	rawTx.IsSubmit = true

	decimals := int32(rawTx.Coin.Contract.Decimals)
	fees := "0"

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
		ExtParam:   rawTx.ExtParam,
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return "", "", nil
}

//CreateSummaryRawTransaction 创建汇总交易
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	var (
		rawTxWithErrArray []*openwallet.RawTransactionWithError
		rawTxArray        = make([]*openwallet.RawTransaction, 0)
		err               error
	)
	rawTxWithErrArray, err = decoder.CreateSummaryRawTransactionWithError(wrapper, sumRawTx)
	if err != nil {
		return nil, err
	}
	for _, rawTxWithErr := range rawTxWithErrArray {
		if rawTxWithErr.Error != nil {
			continue
		}
		rawTxArray = append(rawTxArray, rawTxWithErr.RawTx)
	}
	return rawTxArray, nil
}

//CreateSummaryRawTransactionWithError 创建汇总交易
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		rawTxArray     = make([]*openwallet.RawTransactionWithError, 0)
		accountID      = sumRawTx.Account.AccountID
		accountBalance eos.Asset
		codeAccount    string
		tokenCoin      string
	)

	minTransfer, _ := decimal.NewFromString(sumRawTx.MinTransfer)
	retainedBalance, _ := decimal.NewFromString(sumRawTx.RetainedBalance)

	addr := strings.Split(sumRawTx.Coin.Contract.Address, ":")
	if len(addr) != 2 {
		return nil, fmt.Errorf("token contract's address is invalid: %s", sumRawTx.Coin.Contract.Address)
	}
	codeAccount = addr[0]
	tokenCoin = strings.ToUpper(addr[1])

	if minTransfer.LessThan(retainedBalance) {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	if account.Alias == "" {
		return nil, fmt.Errorf("[%s] have not been created", accountID)
	}

	//账户是否上链
	accountResp, err := decoder.wm.Api.GetAccount(eos.AccountName(account.Alias))
	if err != nil && accountResp == nil {
		return nil, fmt.Errorf("%s account of from not found on chain", decoder.wm.Symbol())
	}

	// 检查目标账户是否存在
	accountTo, err := decoder.wm.Api.GetAccount(eos.AccountName(sumRawTx.SummaryAddress))
	if err != nil && accountTo == nil {
		return nil, fmt.Errorf("%s account of to not found on chain", decoder.wm.Symbol())
	}

	accountAssets, err := decoder.wm.Api.GetCurrencyBalance(eos.AccountName(account.Alias), tokenCoin, eos.AccountName(codeAccount))
	if len(accountAssets) == 0 {
		return rawTxArray, nil
	}

	accountBalance = accountAssets[0]
	accountBalanceDec := decimal.New(int64(accountBalance.Amount), -int32(accountBalance.Precision))

	if accountBalanceDec.LessThan(minTransfer) || accountBalanceDec.LessThanOrEqual(decimal.Zero) {
		return rawTxArray, nil
	}

	//计算汇总数量 = 余额 - 保留余额
	sumAmount := accountBalanceDec.Sub(retainedBalance)

	amountInt64 := sumAmount.Shift(int32(accountBalance.Precision)).IntPart()
	quantity := eos.Asset{Amount: eos.Int64(amountInt64), Symbol: accountBalance.Symbol}
	memo := sumRawTx.GetExtParam().Get("memo").String()

	decoder.wm.Log.Debugf("balance: %v", accountBalanceDec.String())
	decoder.wm.Log.Debugf("fees: %d", 0)
	decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

	//创建一笔交易单
	rawTx := &openwallet.RawTransaction{
		Coin:    sumRawTx.Coin,
		Account: sumRawTx.Account,
		To: map[string]string{
			sumRawTx.SummaryAddress: sumAmount.String(),
		},
		Required: 1,
	}

	createTxErr := decoder.createRawTransaction(
		wrapper,
		rawTx,
		accountResp,
		quantity,
		memo)
	rawTxWithErr := &openwallet.RawTransactionWithError{
		RawTx: rawTx,
		Error: createTxErr,
	}

	//创建成功，添加到队列
	rawTxArray = append(rawTxArray, rawTxWithErr)

	return rawTxArray, nil
}

//createRawTransaction
func (decoder *TransactionDecoder) createRawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	accountResp *eos.AccountResp,
	quantity eos.Asset,
	memo string) *openwallet.Error {

	var (
		to               eos.AccountName
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		keySignList      = make([]*openwallet.KeySignature, 0)
		accountID        = rawTx.Account.AccountID
		amountDec        = decimal.Zero
		codeAccount      = eos.AccountName(rawTx.Coin.Contract.Address)
	)

	for k, v := range rawTx.To {
		to = eos.AccountName(k)
		amountDec, _ = decimal.NewFromString(v)
		break
	}

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(decoder.wm.Api); err != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "filling tx opts: %s", err)
	}

	//action := &eos.Action{
	//	Account: codeAccount,
	//	Name:    token.ActN(decoder.TransferActionName),
	//	Authorization: []eos.PermissionLevel{
	//		{Actor: accountName, Permission: token.PN("active")},
	//	},
	//	ActionData: eos.NewActionData(token.Transfer{
	//		From:     accountName,
	//		To:       to,
	//		Quantity: quantity,
	//		Memo:     memo,
	//	}),
	//}
	action := token.NewTransfer(accountResp.AccountName, to, quantity, memo)
	if codeAccount != action.Account {
		action.Account = codeAccount
	}
	tx := eos.NewTransaction([]*eos.Action{action}, txOpts)
	stx := eos.NewSignedTransaction(tx)
	txdata, cfd, err := stx.PackedTransactionAndCFD()
	if err != nil {
		return openwallet.ConvertError(err)
	}
	//交易哈希
	sigDigest := eos.SigDigest(txOpts.ChainID, txdata, cfd)

	//查找账户的地址，填充待签消息
	for _, permission := range accountResp.Permissions {
		if permission.PermName == "active" {
			for _, pubKey := range permission.RequiredAuth.Keys {
				keyStr, _ := decoder.wm.Decoder.PublicKeyToAddress(pubKey.PublicKey.Content, false)
				addr, err := wrapper.GetAddress(keyStr)
				if err != nil {
					return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "[%s] have not EOS public key: %s", accountID, keyStr)
				}

				signature := openwallet.KeySignature{
					EccType: decoder.wm.Config.CurveType,
					Nonce:   "",
					Address: addr,
					Message: hex.EncodeToString(sigDigest),
				}
				keySignList = append(keySignList, &signature)
			}
		}
	}

	//计算账户的实际转账amount
	//accountTotalSentAddresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", to)
	if accountResp.AccountName != to {
		accountTotalSent = accountTotalSent.Add(amountDec)
	}
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	txFrom = []string{fmt.Sprintf("%s:%s", accountResp.AccountName, amountDec.String())}
	txTo = []string{fmt.Sprintf("%s:%s", to, amountDec.String())}

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	rawTx.RawHex = hex.EncodeToString(txdata)
	rawTx.Signatures[rawTx.Account.AccountID] = keySignList
	rawTx.FeeRate = "0"
	rawTx.Fees = "0"
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.String()
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}
