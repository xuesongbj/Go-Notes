# 单元测试

为测试非导出成员，测试文件也放在目标包内。

&nbsp;

* 测试文件以 `_test.go` 结尾。
    * 通常与测试目标主文件名相同，如 `sort_test.go`。
    * 构建命令(`go build`)忽略测试文件。

&nbsp;

* 测试命令(`go test`):
    * 忽略以 `_` 或 `.` 开头的文件。
    * 忽略 `testdata` 子目录。
    * 执行 `go vet` 检查。

&nbsp;

* 测试函数(`Test<Name>`):
    * `Test` 为识别标记。
    * `<Name>` 为测试名称，首字母大写。如：`TestSort`。

&nbsp;

* 测试函数内以 `Error`、`Fail` 等方法指示测试失败。
    * `Fail`： 失败，继续当前函数。
    * `FailNow`：失败，终止当前函数。
    * `SkipNow`：跳过，终止当前函数。
    * `Log`：输出信息，仅失败或`-v`时有效。
    * `Error`：`Fail + Log`。
    * `Fatal`：`FailNow + Log`。
    * `Skip`： `SkipNow + Log`。

    &nbsp;

    * `os.Exit`：失败，测试进程终止。

&nbsp;

```go
package main

import (
	"testing"
)

func TestAdd(t *testing.T) {
	z := add(1, 2)
	if z != 3 {
		t.FailNow()
	}
}
```

&nbsp;

```bash
root@8d75790f92f5:~/go/david# go test -v
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
ok      david   0.002s
```

&nbsp;

## 模式

* 本地模式(local directory mode)：`go test`，`go test -v`
    * 不缓存测试结果。

* 列表模式(package list mode)：`go test math`，`go test .`， `go test ./...`

    * 缓存结果，直接输出。避免比必要的重复运行。
    * 缓存输出有 `(cached)` 标记。
    * 某些参数(`-count`) 导致缓存失效。
    * 执行 `go clean -testcache` 清理缓存。

```bash
root@8d75790f92f5:~/go/david# go test .
ok      david   (cached)

root@8d75790f92f5:~/go/david# go test -count 2 .
ok      david   0.002s
```

&nbsp;

## 执行

```bash
$ go test           // 测试当前包
$ go test math      // 测试指定包
$ go test ./mylib   // 测试相对路径
$ go test ./...     // 测试当前及所有子包
```

&nbsp;

执行参数：

```go
$> go help test
$> go help testflag
```

&nbsp;

* `-args`：命令行参数。(包列表必须在此参数前)
* `-c`：仅编译，不执行。(用 `-o`修改默认测试可执行文件名)
* `-json`： 输出JSON格式
* `-count`：执行次数。(默认1)
* `-list`：用正则表达式列出测试函数名，不执行
* `-run`：用正则表达式执行要执行的测试函数
* `-timeout`：超时 `panic!`（默认10m）
* `-v`：输出详细信息。
* `-x`：输出构建信息。

&nbsp;

可添加 `test.` 前缀，比如 `-test.v`。(以便和benchmark参数区分)

&nbsp;

```bash
root@8d75790f92f5:~/go/david# go test -v -run "Add" -count 2 -timeout 1m20s ./...
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)

=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
ok      david   0.003s
```
