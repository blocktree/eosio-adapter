package openwtester

import (
	"github.com/blocktree/eosio-adapter/eosio"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
)

func init() {
	//注册钱包管理工具
	log.Notice("Wallet Manager Load Successfully.")
	// openw.RegAssets(eosio.Symbol, eosio.NewWalletManager(nil))

	cache := eosio.NewCacheManager()

	openw.RegAssets(eosio.Symbol, eosio.NewWalletManager(&cache))
}
