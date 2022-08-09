# Delve

Delve是官方扶持的源码级别调试器，能满足日常开发需求。

详细信息请参考[官方文档](https://github.com/go-delve/delve)，本文记录一些基本操作。

* 编译并调试，使用 `dlv debug`。
* 已编译程序，使用 `dlv exec`。
* 已运行程序，使用 `dlv attach`。
* 检查 core dump, 使用 `dlv core`。
* 服务器模式(debug server) 以 `--headless` 启动，客户端 `dlv connect` 接入。

&nbsp;

**调试模式：**

* 自动以 `-gcflags all=-N -l` 方式编译。
* 如想 `go build` 传递参数，使用 `--build-flags`。
* 支持 `runtime.Breakpoint()` 设置断点。

&nbsp;

```go
package main

func add(x, y int) int {
    z := x + y
    return z
}

func main() {
    println(add(1, 2))
}
```

&nbsp;

## 位置

相关位置需要一个位置参数(linespec)。

* `l, ls, list`：查看源码。 </br>

* `<filename>:<line>`：源文件名和行号。
* `<line>`：当前源文件行号。
* `<func>:<line>`：函数内第几行。(定义为起点0)

&nbsp;

* `+<offset>, -<offset>`：基于当前行的偏移量。
* `*<address>`：内存地址。
* `/regex/`：正则表达式匹配的函数。

&nbsp;

```bash
(dlv) b main.add:2
Breakpoint 1 set at 0x45f0ef for main.add() ./main.go:5
(dlv) c
> main.add() ./main.go:5 (hits goroutine(1):1 total:1) (PC: 0x45f0ef)
     1: package main
     2:
     3: func add(x, y int) int {
     4:     z := x + y
=>   5:     return z
     6: }
     7:
     8: func main() {
     9:     println(add(1, 2))
    10: }
```

&nbsp;

## 断点

* `b, break`：设置断点。
* `bp, breakpoints`：显示所有已设置断点。

&nbsp;

* `toggle`：开关断点。
* `clear, clearall`：删除指定或全部断点。

&nbsp;

* `cond, condtion`：设置断点命中条件。
* `on`：设置断点命中执行命令。
* `t, tracepoint`：设置跟踪点。

&nbsp;

```bash
(dlv) b main.add:2
Breakpoint 1 set at 0x45f0ef for main.add() ./main.go:5

(dlv) bp
Breakpoint 1 (enabled) at 0x45f0ef for main.add() ./main.go:5 (0)

(dlv) on 1 locals
(dlv) cond 1 x == 1
(dlv) bp
Breakpoint 1 (enabled) at 0x45f0ef for main.add() ./main.go:5 (0)
    cond x == 1
    locals

(dlv) c
> main.add() ./main.go:5 (hits goroutine(1):1 total:1) (PC: 0x45f0ef)
    z: 3
     1: package main
     2:
     3: func add(x, y int) int {
     4:     z := x + y
=>   5:     return z
     6: }
     7:
     8: func main() {
     9:     println(add(1, 2))
    10: }
```

```bash
(dlv) clear 1
Breakpoint 1 cleared at 0x45f0ef for main.add() ./main.go:5

(dlv) clearall
```

&nbsp;

特殊“断点”(tracepoint)，它不会终端程序执行，仅在命令时(hit) 输出信息。
可用来跟踪代码是否被执行。

```bash
(dlv) t abc main.add
Tracepoint abc set at 0x45f0c0 for main.add() ./main.go:3
```

&nbsp;

可在汇编指令级别设置断点。

```bash
(dlv) b *0x45f120
Breakpoint 2 set at 0x45f120 for main.main() ./main.go:9
(dlv) c
> main.main() ./main.go:9 (hits goroutine(1):1 total:1) (PC: 0x45f120)
     4:     z := x + y
     5:     return z
     6: }
     7:
     8: func main() {
=>   9:     println(add(1, 2))
    10: }
```
