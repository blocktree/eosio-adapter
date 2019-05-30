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
	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
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

func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {

	tokenBalanceList := make([]*openwallet.TokenBalance, 0)

	codeAccount := contract.Address
	tokenCoin := contract.Token

	for _, addr := range address {

		accountAssets, err := decoder.wm.Api.GetCurrencyBalance(eos.AccountName(addr), tokenCoin, eos.AccountName(codeAccount))
		if err != nil {
			decoder.wm.Log.Errorf("get account[%v] token balance failed, err: %v", addr, err)
		}

		if len(accountAssets) == 0 {
			continue
		}

		assets := accountAssets[0]
		accountBalanceDec := decimal.New(int64(assets.Amount), -int32(contract.Decimals))

		tokenBalance := &openwallet.TokenBalance{
			Contract: &contract,
			Balance: &openwallet.Balance{
				Address:          addr,
				Symbol:           contract.Symbol,
				Balance:          accountBalanceDec.String(),
				ConfirmBalance:   accountBalanceDec.String(),
				UnconfirmBalance: "0",
			},
		}

		tokenBalanceList = append(tokenBalanceList, tokenBalance)
	}

	return tokenBalanceList, nil

}

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
		result.ABI = &abiResp.ABI

		if cache != nil {
			cache.Add(keyName, &abiResp.ABI, 0)
		}
	}

	return result, nil
}

// SetABIInfo set abi
func (decoder *ContractDecoder) SetABIInfo(address string, abi openwallet.ABIInfo) error {
	return fmt.Errorf("GetABIInfo not implement")
}
