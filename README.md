# eosio-adapter

## 项目依赖库

- [openwallet](https://github.com/blocktree/openwallet.git)
- [eos-go](https://github.com/eoscanada/eos-go)

## 如何测试

openwtester包下的测试用例已经集成了openwallet钱包体系，创建conf文件，新建EOSIO.ini文件，编辑如下内容：

```ini

# wallet api url
serverAPI = "https://mainnet.eoscanada.com"
# broadcast tx api url
broadcastAPI = "https://mainnet.eoscanada.com"
# Cache data file directory, default = "", current directory: ./data
dataDir = ""

```