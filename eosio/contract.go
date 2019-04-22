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
	"fmt"

	"github.com/blocktree/openwallet/openwallet"
	eos "github.com/eoscanada/eos-go"
)

type ContractDecoder struct {
	openwallet.SmartContractDecoderBase
	wm *WalletManager
}

//NewContractDecoder 智能合约解析器
func NewContractDecoder(wm *WalletManager) *ContractDecoder {
	decoder := ContractDecoder{}
	decoder.wm = wm
	return &decoder
}

//func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {
//
//	codeAccount := contract.Address
//	tokenCoin := contract.Token
//	//tokenDecimals := rawTx.Coin.Contract.Decimals
//
//	//获取wallet
//	account, err := wrapper.GetAssetsAccountInfo(accountID)
//	if err != nil {
//		return err
//	}
//
//	accountAssets, err := decoder.wm.Api.GetCurrencyBalance(eos.AccountName(account.Alias), tokenCoin, eos.AccountName(codeAccount))
//	if len(accountAssets) == 0 {
//		return fmt.Errorf("eos account balance is not enough")
//	}
//
//	accountBalance = accountAssets[0]
//
//}

// GetABIInfo get abi
func (decoder *ContractDecoder) GetABIInfo(address string) (*openwallet.ABIInfo, error) {

	result := &openwallet.ABIInfo{}
	result.Address = address

	keyName := "ABI_" + address

	isCache := false

	cache := decoder.wm.CacheManager
	if cache != nil {
		value, success := cache.Get(keyName)
		if success {
			result.ABI = value
			isCache = true
		}
	}

	if !isCache {
		abiResp, err := decoder.wm.Api.GetABI(eos.AccountName(address))
		if err != nil {
			return nil, fmt.Errorf("get abi from rpc error: %s", err)
		}
		result.ABI = abiResp.ABI

		if cache != nil {
			cache.Add(keyName, abiResp.ABI, 0)
		}
	}

	return result, nil
}

// SetABIInfo set abi
func (decoder *ContractDecoder) SetABIInfo(address string, abi openwallet.ABIInfo) error {
	return fmt.Errorf("GetABIInfo not implement")
}
