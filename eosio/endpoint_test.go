/*
 * Copyright 2019 The openwallet Authors
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
	"github.com/blocktree/openwallet/log"
	"testing"
)

func TestClient_SendRPCRequest(t *testing.T) {
	wm := testNewWalletManager()
	param := map[string]interface{}{
		"account_name": "whaleex.com",
	}
	result, err := wm.client.SendRPCRequest("get_account", param)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("result: %v", string(result))
}
