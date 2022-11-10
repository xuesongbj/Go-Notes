# suffixarray

后缀数组(suffix array) 可用于解决最长公共子串、多模式匹配、最长回文串、全文搜索等问题。</br>
相比于后缀树(suffix tree)，更易于实现，且占用内存更少。

&nbsp;

不过标准库的实现有些简单，仅支持全文检索。

```go
package main

import (
    "fmt"
    "index/suffixarray"
)

func main() {
    s := `abcdefgabxx`
    sub := `ab`

    index := suffixarray.New([]byte(s))
    fmt.Println(index.Lookup([]byte(sub), 10))
}

// [0, 7]
```

&nbsp;

支持正则表达式。和 `Lookup` 不同，`FindAllIndex` 结果是排序过的。

```go
package main

import (
    "fmt"
    "regexp"
    "index/suffixarray"
)

func main() {
    s := `abcdefgabxxs`
    index := suffixarray.New([]byte(s))

    r, _ := regexp.Compile(`ab\w`)
    fmt.Println(index.FindAllIndex(r, -1))
}
```

&nbsp;

## 性能测试

与 `bytes.Index` 实现的正逆向顺序搜索对比。

```go
package main

import (
    "fmt"
    "bytes"
    "index/suffixarray"
)

func bytesIndex(data, sep []byte) []int {
    ret := make([]int, 0, 300)
    pos, length := 0, len(sep)

    for {
        n := bytes.Index(data, sep)
        if n < 0 {
            break
        }
    
        ret = append(ret, pos+n)
        data = data[n+length:]
        pos += (n + length)
    }
    return ret
}

func bytesLastIndex(data, sep []byte) []int {
    ret := make([]int, 0, 300)

    for {
        n := bytes.LastIndex(data, sep)
        if n < 0 {
            break
        }
    
        ret = append(ret, n)
        data = data[:n]
    }

    return ret
}

func main() {
    data := []byte{2, 1, 2, 3, 2, 1, 2, 4, 4, 2, 1}
    sep := []byte{2, 1}

    fmt.Println(bytesIndex(data, sep))
    fmt.Println(bytesLastIndex(data, sep))
    fmt.Println(suffixarray.New(data).Lookup(sep, -1))
}

/*

[0 4 9]
[9 4 0]
[9 0 4]

*/
```

&nbsp;

用一个相对长的源码文件做测试样本数据。

```go
package main

import (
    "testing"
    "io"
    "os"
    "regexp"
    "index/suffixarray"
)

var (
    data, sep []byte
    regx    *regexp.Regexp
    index   *suffixarray.Index
)

func init() {
    f, _ := os.Open("/usr/local/go/src/runtime/malloc.go")
    data, _ = io.ReadAll(f)

    sep = []byte("func")
    regx, _ = regexp.Compile("func")
    index = suffixarray.New(data)
}

func BenchmarkLookup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = index.Lookup(sep, -1)
    }
}

func BenchmarkFindAllIndex(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = index.FindAllIndex(regx, -1)
    }
}

func BenchmarkBytestIndex(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = bytesIndex(data, sep)
    }
}

func BenchmarkBytestLastIndex(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = bytesLastIndex(data, sep)
    }
}
```

&nbsp;

从结果看，`bytes.LastIndex` 最慢, `suffixarray.Lookup` 最快。</br>
付出的代价是，构建后缀数组消耗更多内存。(memprofile)

```bash
root@8d75790f92f5:~/go/david# go test -run None -bench . -benchmem
goos: linux
goarch: amd64
pkg: david
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkLookup-6                3209854           376.0 ns/op       288 B/op          1 allocs/op
BenchmarkFindAllIndex-6           647234          1967 ns/op        1784 B/op          4 allocs/op
BenchmarkBytestIndex-6            107587         11280 ns/op        2688 B/op          1 allocs/op
BenchmarkBytestLastIndex-6         16472         78976 ns/op        2688 B/op          1 allocs/op
PASS
ok      david   6.267s
```
