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
	"encoding/hex"
	"testing"
)

func TestAddressDecoder_PrivateKeyToWIF(t *testing.T) {
	wm := testNewWalletManager()
	prv, _ := hex.DecodeString("")
	wif, _ := wm.Decoder.PrivateKeyToWIF(prv, false)
	t.Logf("wif = %s \n", wif)
}
