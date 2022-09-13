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
