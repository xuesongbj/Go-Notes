# hash

散列函数(hash function) 又称散列算法、哈希函数，为数据创建"指纹"的方法。

&nbsp;

散列函数对数据进行计算，获取称作**散列值**(hash code, hash sum)的指纹，这是数据的简单固定摘要。密码学上，散列算法具有**不可逆性**(无法逆向演算回原来的数值)。将数据运算为另一固定长度值，是其基础原理。

&nbsp;

基本特性: **如果两个散列值不同(同一函数)，那么其原始输入也不相同**。散列函数具有确定性结果，具有这种特性的函数称为单向散列函数。

&nbsp;

另一方面，散列函数的输入和输出不是唯一对应关系。如果两个散列值相同，那么两个输入值既可能相同，也可能不同，此种情况称为"散列碰撞"(collision)。具备抗碰撞(collision-resistant)的算法虽然安全，但计算量大，速度也比较慢。

&nbsp;

散列函数的单项特征，可用来保护数据。比如说，值存储密码哈希值，那么只能验证输入是否正确，而"无法获知原始密码。也可以对比下载的文件摘要，确定是否被篡改。

&nbsp;

## 常见算法

### 消息摘要算法(Message-Digest Algorithm, MD5)

被广泛使用的密码散列函数。可产生128位散列值，确保信息传输完整一致。
后被证实可破解，不适用于高度安全性资料，建议改用SHA-2等。

&nbsp;

### 安全散列算法(Secure Hash Algorithm, SHA)

散列函数家族，是FIPS所认证安全散列算法。计算出数字消息所对应，长度固定的摘要。</br>
美国国家安全局(NSA)设计，由美国国家标准与技术研究院(NIST)发布，是美国政府标准。</br>

&nbsp;

#### SHA-1

1995年发布。广泛用于安全协议(TLS,GnuPG,SSH,IPsec),是MD5后继者。其安全性在2010年后已不被接受。2017年，被正式宣布攻破。

&nbsp;

#### SHA-2

2001年发布，包括SHA-224， SHA-225， SHA-256， SHA-512等。目前没有出现明显弱点。

&nbsp;

#### SHA-3

2015年发布，因MD5被破解，以及SHA-1出现理论破解方法，遂成为替换算法。

&nbsp;

### 基于散列运算的消息认证码(Hash-based Message Authentication Code, HMAC)

利用散列算法，以密钥和消息为输入，生成消息摘要作为输出。</br>
加入了密钥特征的算法，可作加密、数字签名、报文验证等。

&nbsp;

基于密钥的报文完整性验证方法，其安全性是建立在哈希算法的基础上。要求通信双方约定算法、共享密钥，对报文进行运算，形成固定长度认证码。通过认证码校验以确定报文合法性。

&nbsp;

所使用散列函数不限于一种。如用 SHA家族构造的HMAC，被称为 HMAC-SHA等。

&nbsp;

### 非密码学散列函数(FNV)

可快速哈希大量数据并保持较小的冲突概率。</br>

适合一些相近的数据，如IP、URL等。

&nbsp;

### 循环冗余校验(Cyclic Redundancy Check, CRC)

生成固定位数简短校验码的散列函数，用于检查数据传输和存储时可能出现的错误。</br>
不能可靠校验数据完整性(即数据没有发生任何变化)。

&nbsp;

### 校验算法(Adler)

与CRC相比，以可靠性换取速度。

&nbsp;

## Hash源码剖析

```go
type Hash interface {
    // Write (via the embedded io.Writer interface) adds more data to the running hash.
    // It never returns an error.
    io.Writer

    // Sum appends the current hash to b and returns the resulting slice.
    // It does not change the underlying hash state.
    Sum(b []byte) []byte

    // Reset resets the Hash to its initial state.
    Reset()

    // Size returns the number of bytes Sum will return.
    Size() int

    // BlockSize returns the hash's underlying block size.
    // The Write method must be able to accept any amount
    // of data, but it may operate more efficiently if all writes
    // are a multiple of the block size.
    BlockSize() int
}

// Hash32 is the common interface implemented by all 32-bit hash functions.
type Hash32 interface {
    Hash
    Sum32() uint32
}

// Hash64 is the common interface implemented by all 64-bit hash functions.
type Hash64 interface {
    Hash
    Sum64() uint64
}
```

&nbsp;

### 示例

```go
// MD5

package main

import (
    "crypto/md5"
    "fmt"
    "os"
)

func main() {
    b, _ := os.ReadFile("./go.mod")

    h := md5.New()
    h.Write(b)

    fmt.Printf("%x\n", h.Sum(nil))
}
```

```go
//SHA256

package main

import (
    "crypto/sha256"
    "fmt"
)

func main() {
    h := sha256.New()
    h.Write([]byte("hello, world!"))
    fmt.Printf("%x\n", h.Sum(nil))
}
```

```go
// CRC32

package main

import (
    "hash/crc32"
    "fmt"
)

func main() {
    h := crc32.NewIEEE()
    h.Write([]byte("hello, world!"))
    fmt.Printf("%x\n", h.Sum32())
}
```

```go
// HMAC

package main

import (
    "crypto/sha256"
    "crypto/hmac"
    "fmt"
)

func main() {
    key := []byte("password")
    mac := hmac.New(sha256.New, key)
    mac.Write([]byte("hello, world!"))
    fmt.Printf("%x\n", mac.Sum(nil))
}

// 8202be402d11649fc0a00bfd6d6c3ec2017f486fde252e2b504fdd65ea2a29e8
```

&nbsp;

[Online Tools(MD5, SHA256, CRC)](https://emn178.github.io/online-tools/) </br>
[Online HMAC](https://www.freeformatter.com/hmac-generator.html)