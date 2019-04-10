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
	"reflect"
	"testing"

	openwallet "github.com/blocktree/openwallet/openwallet"
	eos "github.com/eoscanada/eos-go"
)

func TestParseHeader(t *testing.T) {
	type args struct {
		b *eos.BlockResp
	}
	tests := []struct {
		name string
		args args
		want *openwallet.BlockHeader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseHeader(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBlock(t *testing.T) {
	type args struct {
		b *eos.BlockResp
	}
	tests := []struct {
		name string
		args args
		want *Block
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseBlock(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
