### 1 从源码编译比特币核心
```

git clone https://github.com/bitcoin/bitcoin.git

cd bitcoin

配置构建比特币核心
For Ubuntu

sudo apt-get install build-essential libtool autotools-dev automake pkg-config libssl-dev libevent-dev bsdmainutils python3 libboost-system-dev libboost-filesystem-dev libboost-chrono-dev libboost-test-dev libboost-thread-dev

自动配置脚本
./autogen.sh

运行configure脚本来自动发现所有必需的库
./configure --prefix=$HOME --disable-wallet --with-incompatible-bdb --with-gui=no

构建Bitcoin核心可执行文件
编译
make

安装
sudo make install

运行比特币核心节点
bitcoind
或者
./bitcoind

```

### 2 配置比特币核心节点
```
查看配置选项
bitcoind --help

设置接受连接的最大节点数
maxconnections

通过删除旧的块，将磁盘空间要求降低到这个兆字节。 但是在丢弃数据之前仍将下载整个数据集。
prune

将交易内存池限制在几兆字节。 使用它来减少节点的内存使用。
maxmempool

在Bitcoin Core配置文件中设置txindex=1。如果不想一开始设置此选项，后期再想设置为完全索引，则需要使用-reindex选项重新启动bitcoind，并等待它重建索引。
维护所有交易的索引
txindex
-txindex=1


设置您将继续的最低费用交易。 低于此值，交易被视为零费用。 在内存受限的节点上使用它来减少内存中交易池的大小。
minrelaytxfee=0.0001

运行
bitcoind -printtoconsole

小型服务器资源不足配置示例 同时会运行
bitcoind -maxconnections=1  -maxmempool=100 -txindex=1 -minrelaytxfee=0.0001 -prune=4000
```
### 3 获得比特币核心相关命令

交易ID在交易被确认之前不具有权威性。在区块链中缺少交易哈希并不意味着交易未被处理。这被称为“交易可扩展性”，因为在块中确认之前可以修改交易哈希。确认后，txid是不可改变的和权威的。


```
客户端状态的信息
bitcoin-cli -getinfo

第1000区块的块哈希值
bitcoin-cli getblockhash 1000
00000000c937983704a73af28acdec37b049d214adbda81d7e2a3dd146f6ed09
```

```
传递交易ID作为参数来检索和检查该交易
bitcoin-cli getrawtransaction 0627052b6f28912f2703066a912ea577f2ce4da4caa5a↵5fbd8a57286c345c2f2
会得到
0100000001186f9f998a5aa6f048e51dd8419a14d8a0f1a8a2836dd734d2804fe65fa35779000↵
000008b483045022100884d142d86652a3f47ba4746ec719bbfbd040a570b1deccbb6498c75c4↵
ae24cb02204b9f039ff08df09cbe9f6addac960298cad530a863ea8f53982c09db8f6e3813014↵
10484ecc0d46f1918b30928fa0e4ed99f16a0fb4fde0735e7ade8416ab9fe423cc54123363767↵
89d172787ec3457eee41c04f4938de5cc17b4a10fa336a8d752adfffffffff0260e3160000000↵
0001976a914ab68025513c3dbd2f7b92a94e0581f5d50f654e788acd0ef8000000000001976a9↵
147f9b1a7fb68d60c536c2fd8aeaa53a8f3cc025a888ac00000000

使用decodeawtransaction命令
bitcoin-cli decoderawtransaction 0100000001186f9f998a5aa6f048e51dd8419a14d8↵
a0f1a8a2836dd734d2804fe65fa35779000000008b483045022100884d142d86652a3f47ba474↵
6ec719bbfbd040a570b1deccbb6498c75c4ae24cb02204b9f039ff08df09cbe9f6addac960298↵
cad530a863ea8f53982c09db8f6e381301410484ecc0d46f1918b30928fa0e4ed99f16a0fb4fd↵
e0735e7ade8416ab9fe423cc5412336376789d172787ec3457eee41c04f4938de5cc17b4a10fa↵
336a8d752adfffffffff0260e31600000000001976a914ab68025513c3dbd2f7b92a94e0581f5↵
d50f654e788acd0ef8000000000001976a9147f9b1a7fb68d60c536c2fd8aeaa53a8f3cc025a8↵
88ac00000000

得到交易详情的json字符串
{
  "txid": "0627052b6f28912f2703066a912ea577f2ce4da4caa5a5fbd8a57286c345c2f2",
  "size": 258,
  "version": 1,
  "locktime": 0,
  "vin": [
    {
      "txid": "7957a35fe64f80d234d76d83a2...8149a41d81de548f0a65a8a999f6f18",
      "vout": 0,
      "scriptSig": {
        "asm":"3045022100884d142d86652a3f47ba4746ec719bbfbd040a570b1decc...",
        "hex":"483045022100884d142d86652a3f47ba4746ec719bbfbd040a570b1de..."
      },
      "sequence": 4294967295
    }
  ],
  "vout": [
    {
      "value": 0.01500000,
      "n": 0,
      "scriptPubKey": {
        "asm": "OP_DUP OP_HASH160 ab68...5f654e7 OP_EQUALVERIFY OP_CHECKSIG",
        "hex": "76a914ab68025513c3dbd2f7b92a94e0581f5d50f654e788ac",
        "reqSigs": 1,
        "type": "pubkeyhash",
        "addresses": [
          "1GdK9UzpHBzqzX2A9JFP3Di4weBwqgmoQA"
        ]
      }
    },
    {
      "value": 0.08450000,
      "n": 1,
      "scriptPubKey": {
        "asm": "OP_DUP OP_HASH160 7f9b1a...025a8 OP_EQUALVERIFY OP_CHECKSIG",
        "hex": "76a9147f9b1a7fb68d60c536c2fd8aeaa53a8f3cc025a888ac",
        "reqSigs": 1,
        "type": "pubkeyhash",
        "addresses": [
          "1Cdid9KFAaatwczBwBttQcwXYCpvK8h7FK"
        ]
      }
    }
  ]
}
```
本次交易的txid索引的前一笔交易进一步探索区块链.
```
使用getblock命令，并把区块哈希值作为参数来查询对应的区块
bitcoin-cli getblockhash 277316
0000000000000001b6b9a13b095e96db41c4a928b97ef2d944a9b31b2cc7bdc4

区块哈希值作为参数来查询对应的区块
bitcoin-cli getblock 0000000000000001b6b9a13b095e96db41c4a928b97ef2d944a9b3↵
1b2cc7bdc4

{
  "hash": "0000000000000001b6b9a13b095e96db41c4a928b97ef2d944a9b31b2cc7bdc4",
  "confirmations": 37371,
  "size": 218629,
  "height": 277316,
  "version": 2,
  "merkleroot": "c91c008c26e50763e9f548bb8b2fc323735f73577effbc55502c51eb4cc7cf2e",
  "tx": [
    "d5ada064c6417ca25c4308bd158c34b77e1c0eca2a73cda16c737e7424afba2f",
    "b268b45c59b39d759614757718b9918caf0ba9d97c56f3b91956ff877c503fbe",
    "04905ff987ddd4cfe603b03cfb7ca50ee81d89d1f8f5f265c38f763eea4a21fd",
    "32467aab5d04f51940075055c2f20bbd1195727c961431bf0aff8443f9710f81",
    "561c5216944e21fa29dd12aaa1a45e3397f9c0d888359cb05e1f79fe73da37bd",
[... hundreds of transactions ...]
    "78b300b2a1d2d9449b58db7bc71c3884d6e0579617e0da4991b9734cef7ab23a",
    "6c87130ec283ab4c2c493b190c20de4b28ff3caf72d16ffa1ce3e96f2069aca9",
    "6f423dbc3636ef193fd8898dfdf7621dcade1bbe509e963ffbff91f696d81a62",
    "802ba8b2adabc5796a9471f25b02ae6aeee2439c679a5c33c4bbcee97e081196",
    "eaaf6a048588d9ad4d1c092539bd571dd8af30635c152a3b0e8b611e67d1a1af",
    "e67abc6bd5e2cac169821afc51b207127f42b92a841e976f9b752157879ba8bd",
    "d38985a6a1bfd35037cb7776b2dc86797abbb7a06630f5d03df2785d50d5a2ac",
    "45ea0a3f6016d2bb90ab92c34a7aac9767671a8a84b9bcce6c019e60197c134b",
    "c098445d748ced5f178ef2ff96f2758cbec9eb32cb0fc65db313bcac1d3bc98f"
  ],
  "time": 1388185914,
  "mediantime": 1388183675,
  "nonce": 924591752,
  "bits": "1903a30c",
  "difficulty": 1180923195.258026,
  "chainwork": "000000000000000000000000000000000000000000000934695e92aaf53afa1a",
  "previousblockhash": "0000000000000002a7bbd25a417c0374cc55261021e8a9ca74442b01284f0569",
  "nextblockhash": "000000000000000010236c269dd6ed714dd5db39d36b33959079d78dfd431ba7"
}

```
### 4 使用比特币核心的编程接口
bitcoin-cli helper对于探索Bitcoin Core API和测试功能非常有用。但是应用编程接口的全部要点是以编程方式访问功能。
 JSON代表JavaScript对象符号，RPC代表远程过程调用
RESTful
 命令行HTTP客户端来构造这些JSON-RPC调用之一：
 ```
  curl --user myusername --data-binary '{"jsonrpc": "1.0", "id":"curltest", "method": "getinfo", "params": [] }' -H 'content-type: text/plain;' http://127.0.0.1:8332/
 ```
 
 
 
 ### 5 其他替代客户端、资料库、工具包
 C/C++

[Bitcoin Core](https://github.com/bitcoin/bitcoin)    The reference implementation of bitcoin

[libbitcoin](https://github.com/libbitcoin/libbitcoin)    Cross-platform C++ development toolkit, node, and consensus library

[bitcoin explorer](https://github.com/libbitcoin/libbitcoin-explorer)    Libbitcoin’s command-line tool

[picocoin](https://github.com/jgarzik/picocoin)    A C language lightweight client library for bitcoin by Jeff Garzik

JavaScript

[bcoin](http://bcoin.io/)    A modular and scalable full-node implementation with API

[Bitcore](https://bitcore.io/)    Full node, API, and library by Bitpay

[BitcoinJS](https://github.com/bitcoinjs/bitcoinjs-lib)    A pure JavaScript Bitcoin library for node.js and browsers

Java

[bitcoinj](https://bitcoinj.github.io/)    A Java full-node client library

[Bits of Proof \(BOP\)](https://bitsofproof.com/)    A Java enterprise-class implementation of bitcoin

Python

[python-bitcoinlib](https://github.com/petertodd/python-bitcoinlib)    A Python bitcoin library, consensus library, and node by Peter Todd

[pycoin](https://github.com/richardkiss/pycoin)    A Python bitcoin library by Richard Kiss

[pybitcointools](https://github.com/vbuterin/pybitcointools)    A Python bitcoin library by Vitalik Buterin

Ruby

[bitcoin-client](https://github.com/sinisterchipmunk/bitcoin-client)    A Ruby library wrapper for the JSON-RPC API

Go

[btcd](https://github.com/btcsuite/btcd)    A Go language full-node bitcoin client

Rust

[rust-bitcoin](https://github.com/apoelstra/rust-bitcoin)    Rust bitcoin library for serialization, parsing, and API calls

C\#

[NBitcoin](https://github.com/MetacoSA/NBitcoin)    Comprehensive bitcoin library for the .NET framework

Objective-C

[CoreBitcoin](https://github.com/oleganza/CoreBitcoin)    Bitcoin toolkit for ObjC and Swift


# 密钥和地址
比特币的所有权是通过数字密钥、比特币地址和数字签名来确定的。数字密钥实际上并不存储在网络中，而是由用户生成之后，存储在一个叫做钱包的文件或简单的数据库中。

### 素数幂和椭圆曲线乘法。这些数学函数都是不可逆的
使得生成数字密钥和不可伪造的数字签名成为可能。比特币正是使用椭圆曲线乘法作为其公钥加密的基础。
在比特币系统中，我们用公钥加密创建一个密钥对，用于控制比特币的获取。
密钥对包括一个私钥，和由其衍生出的唯一的公钥。
公钥用于接收比特币，而私钥用于比特币支付时的交易签名。

公钥和私钥之间的数学关系，使得私钥可用于生成特定消息的签名。
此签名可以在不泄露私钥的同时对公钥进行验证。
支付比特币时，比特币的当前所有者需要在交易中提交其公钥和签名（每次交易的签名都不同，但均从同一个私钥生成）。比特币网络中的所有人都可以通过所提交的公钥和签名进行验证，并确认该交易是否有效，即确认支付者在该时刻对所交易的比特币拥有所有权

非对称密码学的有用属性是生成数字签名的能力


## 公钥
通过椭圆曲线乘法可以从私钥计算得到公钥，这是不可逆转的过程：K = k * G。其中k是私钥，G是被称为生成点的常数点，而K是所得公钥。
 椭圆曲线乘法是在一个方向（乘法）很容易做，而不可能在相反的方向（除法）做。
 
 
 下图显示了在曲线上得到`G、2G、4G`的几何操作。
 ![图4-4曲线上 G、2G、4G 的几何操作](http://upload-images.jianshu.io/upload_images/1785959-52ef6b8a628405a8.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
 给定椭圆曲线上的两个点P1和P2，则椭圆曲线上必定有第三点 P3 = P1 + P2。

几何图形中，该第三点P3可以在P1和P2之间画一条线来确定。这条直线恰好与椭圆曲线相交于另外一个地方。此点记为P3'= (x，y)。然后，在x轴做翻折获得P3=(x，-y)。

