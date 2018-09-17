### 比特币地址
公式中，K是公钥，A是生成的比特币地址。
A = RIPEMD160(SHA256(K))

![图4-5从公钥生成比特币地址](http://upload-images.jianshu.io/upload_images/1785959-6fc43eee55666ff2.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

### 公钥
分为非压缩格式或压缩格式公钥这两种形式
### 压缩格式公钥 
公钥长度变长了，但可以依据椭圆曲线特性，找到另一个点，可以只存y轴的点坐标

引入压缩格式公钥是为了减少比特币交易的字节数，从而可以节省那些运行区块链数据库的节点磁盘空间。

下图阐释了公钥压缩：

![图4-7公钥压缩](http://upload-images.jianshu.io/upload_images/1785959-4e4e3255fc54395b.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

未压缩格式公钥使用04作为前缀，而压缩格式公钥是以02或03作为前缀。

为了区分y坐标的两种可能值，我们在生成压缩格式公钥时，如果y是偶数，则使用02作为前缀；如果y是奇数，则使用03作为前缀。这样就可以根据公钥中给定的x值，正确推导出对应的y坐标，从而将公钥解压缩为在椭圆曲线上的完整的点坐标。

###  Base58和Base58Check编码

Base58不含Base64中的0（数字0）、O（大写字母o）、l（小写字母L）、I（大写字母i），以及“+”和“/”两个字符。

校验和是从编码的数据的哈希值中得到的，所以可以用来检测并避免转录和输入中产生的错误。

使用Base58check编码时，解码软件会计算数据的校验和并和编码中自带的校验和进行对比。

一个错误比特币地址就不会被钱包软件认为是有效的地址，否则这种错误会造成资金的丢失。

checksum = SHA256(SHA256(prefix+data))

![图4-6Base58Check编码的过程](http://upload-images.jianshu.io/upload_images/1785959-fd3d820e5ba1474c.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)


在比特币中，大多数需要向用户展示的数据都使用Base58Check编码，可以实现数据压缩，易读而且有错误检验。

Base58Check编码的私钥WIF是以5开头的。

表4-1 Base58Check版本前缀和编码后的结果


![Base58Check版本前缀和编码后的结果](http://upload-images.jianshu.io/upload_images/1785959-f3cd346fa1b81b2b.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

### 密钥的格式

#### 私钥的格式

私钥的三种常见格式
![表4-2展私钥的三种常见格式](http://upload-images.jianshu.io/upload_images/1785959-fa916ebbf5c0d2cf.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

示例：同样的私钥，不同的格式

![表4-3 示例：同样的私钥，不同的格式](http://upload-images.jianshu.io/upload_images/1785959-1654059037c90c8b.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

虽然编码后的字符串看起来不同，但不同的格式彼此之间可以很容易地相互转换。

使用Bitcoin Explorer中的wif-to-ec命令来显示两个WIF键代表相同的私钥：

```sh
$ bx wif-to-ec 5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn

1e99423a4ed27608a15a2616a2b0e9e52ced330ac530edcc32c8ffc6a526aedd

$ bx wif-to-ec KxFC1jmwwCoACiCAWZ3eXa96mBM6tb3TYzGmf6YwgdGWZgawvrtJ

1e99423a4ed27608a15a2616a2b0e9e52ced330ac530edcc32c8ffc6a526aedd
```

##### 从Base58Check解码
```
$ bx base58check-decode 5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn

wrapper

{
...
checksum 4286807748
payload 1e99423a4ed27608a15a2616a2b0e9e52ced330ac530edcc32c8ffc6a526aedd
version 128
...
}
```
结果包含密钥作为有效载荷，WIF版本前缀128和校验和。
##### 将十六进制转换为Base58Check编码
##### 将十六进制（压缩格式密钥）转换为Base58Check编码

### 压缩格式私钥
当一个私钥被使用WIF压缩格式导出时，不但没有压缩，而且比“非压缩格式”私钥长出一个字节。
用以表明该私钥是来自于一个较新的钱包，只能被用来生成压缩的公钥。私钥是非压缩的，也不能被压缩。“压缩的私钥”实际上只是表示“用于生成压缩格式公钥的私钥”，而“非压缩格式私钥”用来表明“用于生成非压缩格式公钥的私钥”。

表4示例：相同的密钥，不同的格式

![表4-4展示了同样的私钥使用不同的WIF和WIF压缩格式编码。](http://upload-images.jianshu.io/upload_images/1785959-9b55c974726bc841.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

这些格式并不是可互换使用的。在实现了压缩格式公钥的较新的钱包中，私钥只能且永远被导出为WIF压缩格式（以K或L为前缀）。对于较老的没有实现压缩格式公钥的钱包，私钥将只能被导出为WIF格式（以5为前缀）导出。这样做的目的就是为了给导入这些私钥的钱包一个信号：是否钱包必须搜索区块链寻找压缩或非压缩公钥和地址。

如果一个比特币钱包实现了压缩格式公钥，那么它将会在所有交易中使用该压格式缩公钥。钱包中的私钥将会被用来在曲线上生成公钥点，这个公钥点将会被压缩。压缩格式公钥然后被用来生成交易中使用的比特币地址。当从一个实现了压缩格式公钥的新的比特币钱包导出私钥时，钱包导入格式（WIF）将会被修改为WIF压缩格式，该格式将会在私钥的后面附加一个字节大小的后缀01。最终的Base58Check编码格式的私钥被称作WIF（“压缩”）私钥，以字母“K”或“L”开头。而以“5”开头的是从较老的钱包中以WIF（非压缩）格式导出的私钥。

### 高级密钥和地址

#### 加密私钥（BIP0038）
BIP0038提出了一个通用标准，使用一个口令加密私钥并使用Base58Check对加密的私钥进行编码，这样加密的私钥就可以安全地保存在备份介质里，安全地在钱包间传输，保持密钥在任何可能被暴露情况下的安全性。这个加密标准使用了AES。

BIP0038加密方案是：输入一个比特币私钥，通常使用WIF编码过，base58chek字符串的前缀“5”。此外BIP0038加密方案需要一个长密码作为口令，通常由多个单词或一段复杂的数字字母字符串组成。BIP0038加密方案的结果是一个由base58check编码过的加密私钥，前缀为6P。


###  P2SH (Pay-to-Script Hash)和多重签名地址

P2SH函数最常见的实现是多重签名地址脚本。顾名思义，底层脚本需要多个签名来证明所有权，此后才能消费资金。

script hash = RIPEMD160(SHA256(script))

比特币靓号地址
靓号地址安全性
靓号地址也可能使得任何人都能创建一个类似于随机地址的地址，甚至另一个靓号地址，从而欺骗你的客户。
反之 
生成8个字符的靓号地址，攻击者将会被逼迫到10字符的境地

#### 纸钱包
一个更复杂的纸钱包存储系统使用BIP0038加密的私钥。打印在纸钱包上的这些私钥被其所有者记住的一个口令保护起来。没有口令，这些被加密过的密钥也是毫无用处的。