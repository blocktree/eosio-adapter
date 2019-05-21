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
		accountBalanceDec := decimal.New(int64(assets.Amount), -int32(assets.Precision))

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