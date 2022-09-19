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

&nbsp;

## 并行

默认情况下，包内串行，多包并行。

```go
// main_test.go

package main

import (
    "testing"
    "time"
)

func TestA(t *testing.T){
    time.Sleep(time.Second * 10)
}

func TestB(t *testing.T) {
    time.Sleep(time.Second * 10)
}
```

&nbsp;

```go
// ./mylib/demo_test.go

package mylib

import (
    "testing"
    "time"
)

func TestX(t *testing.T) {
    time.Sleep(time.Second * 10)
}
```

&nbsp;

```bash
# 单个包，串行。

root@8d75790f92f5:~/go/david# go clean -cache -testcache
root@8d75790f92f5:~/go/david# time go test -v
=== RUN   TestA
--- PASS: TestA (10.00s)
=== RUN   TestB
--- PASS: TestB (10.01s)
PASS
ok      david   20.014s

real    0m21.084s       <!----- !!!
user    0m1.654s
sys 0m1.280s
```

&nbsp;

```bash
# 多个包，并行

root@8d75790f92f5:~/go/david# go clean -cache -testcache
root@8d75790f92f5:~/go/david# time go test -v -run "Test[AX]" ./...
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestA
--- PASS: TestA (10.01s)
PASS
ok      david   10.013s
=== RUN   TestX
--- PASS: TestX (10.01s)
PASS
ok      david/mylib 10.013s

real    0m11.150s   <!----- !!!
user    0m1.896s
sys 0m1.479s
```

参数 `-p, -parallel` 用于设置 `GOMAXPROCS`, 这会影响并发执行。

&nbsp;

```bash
root@8d75790f92f5:~/go/david# go clean -cache -testcache
root@8d75790f92f5:~/go/david# time go test -v -run "Test[AX]" -p 1 ./...
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestA
--- PASS: TestA (10.01s)
PASS
ok      david   10.014s
=== RUN   TestX
--- PASS: TestX (10.01s)
PASS
ok      david/mylib 10.012s

real    0m22.242s
user    0m1.680s
sys 0m1.448s
```

&nbsp;

调用 `Parallel` 后，测试函数暂停(`PAUSE`)。</br>
等所有串行测试结束后，再恢复(`CONT`)并发执行。

&nbsp;

```go
package main

import (
    "testing"
    "time"
)

func TestA(t *testing.T){
    t.Parallel()
    time.Sleep(time.Second * 10)
}

func TestB(t *testing.T) {
    t.Parallel()
    time.Sleep(time.Second * 10)
}

func TestC(t *testing.T) {
    t.Parallel()
    time.Sleep(time.Second * 10)
}
```

&nbsp;

```bash
# AB并行
root@8d75790f92f5:~/go/david# go clean -cache -testcache; go test -v -run "[AB]"
=== RUN   TestA
=== PAUSE TestA
=== RUN   TestB
=== PAUSE TestB
=== CONT  TestA
=== CONT  TestB
--- PASS: TestB (10.00s)
--- PASS: TestA (10.00s)
PASS
ok      david   10.009s
```

&nbsp;

```bash
# AB并行，C 串行
root@8d75790f92f5:~/go/david# go clean -cache -testcache; go test -v
=== RUN   TestA
=== PAUSE TestA
=== RUN   TestB
=== PAUSE TestB
=== RUN   TestC
=== PAUSE TestC
=== CONT  TestA
=== CONT  TestC
=== CONT  TestB
--- PASS: TestB (10.01s)
--- PASS: TestA (10.01s)
--- PASS: TestC (10.01s)
PASS
ok      david   10.013s
```

&nbsp;

如设置 `-cpu`、`-count` 参数，那么同一测试函数依旧串行执行多次。

```bash
root@8d75790f92f5:~/go/david# go clean -cache -testcache; go test -v -count 2 -run "TestA"
=== RUN   TestA
=== PAUSE TestA
=== CONT  TestA
--- PASS: TestA (10.00s)
=== RUN   TestA
=== PAUSE TestA
=== CONT  TestA
--- PASS: TestA (10.00s)
PASS
ok      david   20.009s
```
