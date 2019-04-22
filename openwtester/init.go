package openwtester

import (
	"time"

	"github.com/blocktree/eosio-adapter/eosio"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
	"github.com/blocktree/openwallet/openwallet"
)

type CacheManager struct {
	bucket map[string]interface{}
}

func (cm *CacheManager) Add(key string, value interface{}, duration time.Duration) error {
	cm.bucket[key] = value
	return nil
}
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	value := cm.bucket[key]
	if value == nil {
		return nil, false
	}
	return value, true
}
func (cm *CacheManager) GetCacheEntry(key string) (*openwallet.CacheEntry, bool) {
	return nil, true
}
func (cm *CacheManager) Remove(key string) (interface{}, bool) {
	return nil, true
}
func (cm *CacheManager) Contains(key string) bool {
	return true
}
func (cm *CacheManager) Clear() {}

func init() {
	//注册钱包管理工具
	log.Notice("Wallet Manager Load Successfully.")
	// openw.RegAssets(eosio.Symbol, eosio.NewWalletManager(nil))

	cache := &CacheManager{}
	cache.bucket = make(map[string]interface{})
	openw.RegAssets(eosio.Symbol, eosio.NewWalletManager(cache))
}
