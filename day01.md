# 1.1 比特币入门
数字货币生态系统基础的概念和技术的集合

* 去中心化的对等网络（比特币协议）

* 公共交易总帐（区块链）

* 独立交易确认和货币发行的一套规则（共识规则）

* 实现有效的区块链全球去中心化共识的机制（工作量证明算法）

开发人员，可以将比特币视为货币互联网，即通过分布式计算传播价值和确保数字资产所有权的网络

**1.2 比特币之前的数字货币**

比特币解决了什么问题

数字货币有三个基本问题：

* 真假 (人人相信，比特币系统的信任是建立在计算的基础上的)
* 双花 (分布式网络达成关于交易状态的共识)
* 货币所有权 (私钥)

**1.3 历史**
* b-money和HashCash
* 2008 Satoshi Nakamoto(中本聪) Bitcoin：A Peer-to-Peer Electronic Cash System
* 2009,中本聪编写 比特币客户端
* 拜占庭将军问题 同名论文中提出的**分布式对等网络通信容错问题**
* 在分布式计算中，不同的计算机通过通讯交换信息达成共识而按照同一套协作策略行动。
* 中本聪的解决方案使用工作量证明算法在没有中央信任机构的情况下实现共识，代表了分布式计算的突破，并具有超越数字货币应用的广泛适用性。
***
# 2.1 交易，区块，挖矿和区块链
![](http://upload-images.jianshu.io/upload_images/1785959-cf934db5a82e4f15.png?imageMogr2/auto-orient/strip|imageView2/2/w/1240)
# 2.2 比特币交易
# 查询交易常用网站
>[Bitcoin Block Explorer](https://blockexplorer.com/)

>[BlockCypher Explorer](https://live.blockcypher.com/)

>[blockchain.info](https://blockchain.info/)

>[BitPay Insight](https://insight.bitpay.com/)

交易告知全网：比特币的持有者已授权把比特币转帐给其他人

“消费”指的是签署一笔交易：转移一笔以前交易的比特币给比特币地址所标识的新所有者

交易链  交易形成了一条链，最近交易的输入对应以前交易的输出。

交易的构建

钱包应用甚至可以在完全离线时建立交易，比特币交易建立和签名时不用连接比特币网络，只有在执行交易时才需要将交易发送到网络。


![](http://upload-images.jianshu.io/upload_images/1785959-d646668b27410d82.png?imageMogr2/auto-orient/strip|imageView2/2/w/1240)

bob's coffee 支付请求二维码

```
bitcoin:1GdK9UzpHBzqzX2A9JFP3Di4weBwqgmoQA?

amount=0.015&
label=Bob%27s%20Cafe&
message=Purchase%20at%20Bob%27s%20Cafe

Components of the URL

A bitcoin address: "1GdK9UzpHBzqzX2A9JFP3Di4weBwqgmoQA"
The payment amount: "0.015"
A label for the recipient address: "Bob's Cafe"
A description for the payment: "Purchase at Bob's Cafe"

```



千分之一比特币(1毫比特币）

一亿分之一比特币（1聪比特币）

# 2.3 区块

# UTXO

```
{
    "unspent_outputs":[

        {
            "tx_hash":"186f9f998a5...2836dd734d2804fe65fa35779",
            "tx_index":104810202,
            "tx_output_n": 0,
            "script":"76a9147f9b1a7fb68d60c536c2fd8aeaa53a8f3cc025a888ac",
            "value": 10000000,
            "value_hex": "00989680",
            "confirmations":0
        }

    ]
}
```
# 2.4 挖矿
挖矿在比特币系统中有两个重要作用
* 挖矿节点通过参考比特币的共识规则验证所有交易。挖矿通过拒绝无效或畸形交易来提供比特币交易的安全性。

* 挖矿在构建区块时会创造新的比特币,每个区块创造的比特币数量是固定的，随时间会渐渐减少。

特点

它解起来困难而验证很容易，并且它的困难度可以调整。类似数独游戏。