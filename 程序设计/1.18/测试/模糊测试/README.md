# 模糊测试

又称随机测试，一种基于随机输入发现代码缺陷的自动化测试技术。

&nbsp;

* 对单元测试预定数据的补充，查找未预料错误。
* 并非验证逻辑正确，而是发现输入处理缺陷。

&nbsp;

* 初始数据，称作**种子语料(seed corpus)**。
* 基于种子构造的随机数据，称作 **随机语料(random corpus)**。

```go
// mylib.go

package mylib

func add(a, b int) int {
    return a + b
}
```

```go
package mylib

import (
    "testing"
)


func FuzzAdd(f *testing.F) {
    // 添加种子.(可选)
    f.Add(1, 1)
    f.Add(1, 2)
    f.Add(1, 3)

    // 随机测试。(第一参数为*T, 后续和测试目标函数相同)
    f.Fuzz(func(t *testing.T, x, y int) {
        add(x, y)
    })
}
```

&nbsp;

默认被当作普通单元测试运行。测试数据就是种子，不会引入随机语料。

```bash
root@8d75790f92f5:~/go/david# go test -v ./mylib
=== RUN   FuzzAdd
=== RUN   FuzzAdd/seed#0
=== RUN   FuzzAdd/seed#1
=== RUN   FuzzAdd/seed#2
--- PASS: FuzzAdd (0.00s)
    --- PASS: FuzzAdd/seed#0 (0.00s)
    --- PASS: FuzzAdd/seed#1 (0.00s)
    --- PASS: FuzzAdd/seed#2 (0.00s)
PASS
ok      david/mylib 0.003s
```

&nbsp;

只有添加 `-fuzz` 参数才会被进行随机测试。</br>
默认无限期执行，直到失败或被用户中断(`CTRL+C`)。

* `-fuzz`: 测试目标。(regex)
* `-fuzztime`：时长或次数。(`1m20s`, `100x`)

&nbsp;

```go
root@8d75790f92f5:~/go/david# go test -v -fuzz Add -fuzztime 20s ./mylib
=== FUZZ  FuzzAdd
fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 6 workers
fuzz: elapsed: 3s, execs: 314171 (104687/sec), new interesting: 0 (total: 3)
fuzz: elapsed: 6s, execs: 632725 (106187/sec), new interesting: 0 (total: 3)
fuzz: elapsed: 9s, execs: 902351 (89887/sec), new interesting: 0 (total: 3)
fuzz: elapsed: 12s, execs: 1218743 (105463/sec), new interesting: 0 (total: 3)
fuzz: elapsed: 15s, execs: 1526489 (102591/sec), new interesting: 0 (total: 3)
fuzz: elapsed: 18s, execs: 1854967 (109491/sec), new interesting: 0 (total: 3)
fuzz: elapsed: 20s, execs: 2043246 (90113/sec), new interesting: 0 (total: 3)
--- PASS: FuzzAdd (20.09s)
PASS
ok      david/mylib 20.095s
```

&nbsp;

基线覆盖率(baseline coverage)是对现有语料(种子等)的测试结果，提供基准指标。</br>
执行期间，如某条随机输入导致语料库之外的覆盖变化，那么可称作“new interesting”。

&nbsp;

> "`new interesting`" 保存在 `GOCACHE/fuzz` 目录下，可被 `go clean -fuzzcache` 清除。

&nbsp;

* `elapsed`：启动时间。
* `execs`：模糊输入总数。
* `new interesting`：引发覆盖率变化的输入数，以及语料总数。

&nbsp;

* 引发失败的原因:
    * `panic`
    * `t.Fail...`
    * `os.Exit`, `stack overflow` ...
    * 目标执行时间过长 (默认1秒)

&nbsp;

## 预料库

除了在代码中添加种子外，还可以将其存储在文件内，自动载入。

* 路径: `testdata/fuzz/<FuzzzName>/`
* 文件: 每个文件保存一组语料，也就是一次`Add` 调用的参数。

```bash
root@8d75790f92f5:~/go/david/mylib# cat testdata/fuzz/FuzzAdd/a
go test fuzz v1
int(1)
int(2)

root@8d75790f92f5:~/go/david/mylib# cat testdata/fuzz/FuzzAdd/b
go test fuzz v1
int(1)
int(2)

root@8d75790f92f5:~/go/david/mylib# cat testdata/fuzz/FuzzAdd/c
go test fuzz v1
int(1)
int(3)
```

测试出错时，也会将随机输入存储到该路径下。
作为回归测试的基准语料，以检查目标是否被修复。

&nbsp;

[Go Fuzzing](https://go.dev/security/fuzz/) </br>
[Fuzzing support for Go](https://www.yuque.com/r/goto?url=https%3A%2F%2Fdocs.google.com%2Fdocument%2Fu%2F1%2Fd%2F1zXR-TFL3BfnceEAWytV8bnzB2Tfp6EPFinWVJ5V4QC8%2Fpub)