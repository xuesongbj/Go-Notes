# 编译

编译不仅仅进行是执行 `go build` 命令，还有些额外内容。

* **调试**: 参数 `-gcflags "-N -l"`，阻止优化和内联。
* **发布**: 参数 `-ldflags "-w -s"`，剔除符号表和调试信息。
* **保护**: 借助专业工具(UPX)，对可执行文件进行减肥和保护。

&nbsp;

## 编译指令

编译命令(#pragmas) 非语言规范内容，由编译器实现，指示对代码做"特殊"处理。

* `//go:noinline`： 阻止内联。
* `//go:nosplit`：不插入 `morestack` 指令。(栈检查和扩容)

&nbsp;

> 注意, `//go` 之间不能有空格，避免和注释(comment)冲突。</br>
> 使用编译指令阻止具体某个函数内联，这与 `-gcflags "-l"` 全局不同。

&nbsp;

```go
package main

//go:noinline
//go:nosplit
func test() {
    println("hello, world!")
}

func main() {
    test();
}
```

```bash
$ go build -o test
$ go tool objdump -s "main\.test" ./test

TEXT main.test(SB)
  main.go:5     SUBQ $0x18, SP
  main.go:5     MOVQ BP, 0x10(SP)
  main.go:5     LEAQ 0x10(SP), BP
  
  main.go:6     CALL runtime.printlock(SB)
  main.go:6     LEAQ 0xd7a4(IP), AX
  main.go:6     MOVL $0xe, BX
  main.go:6     NOPL
  main.go:6     CALL runtime.printstring(SB)
  main.go:6     CALL runtime.printunlock(SB)
  
  main.go:7     MOVQ 0x10(SP), BP
  main.go:7     ADDQ $0x18, SP
  main.go:7     RET
```

&nbsp;

## 编译时间

不修改源码，动态调整某些内容。比如编译时间、自增版本号等。

```go
// test/lib/lib.go

package lib

var buildTime string
```

```bash
$> go build -ldflags "-X test/mylib.ver=$(date +'%Y%m%d-%H%M%S')"
```

&nbsp;

## 类型信息

编译器将类型信息存储到`.rodata`段，以 `lea` 指令加载。

```go
package main

type data struct {
    x int
}

func main() {
    m := make(map[string]data)
    m["a"] = data{}
}
```

&nbsp;

```bash
$ go build -gcflags -S 2>a.txt     # 输出到文件，便于查看。

"".main STEXT size=170 args=0x0 locals=0x128 funcid=0x0 align=0x0

    LEAQ    type.map[string]"".data(SB), AX
    ...

type.map[string]"".data SRODATA dupok size=88
    0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00
    0x0010 65 42 fb e0 02 08 08 35 00 00 00 00 00 00 00 00
    0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
    0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
    0x0040 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
    0x0050 10 08 d0 00 0c 00 00 00                        
    rel 32+8 t=1 runtime.gcbits.01+0
    rel 40+4 t=5 type..namedata.*map[string]main.data-+0
    rel 44+4 t=-32763 type.*map[string]"".data+0
    rel 48+8 t=1 type.string+0
    rel 56+8 t=1 type."".data+0
    rel 64+8 t=1 type.noalg.map.bucket[string]"".data+0
    rel 72+8 t=1 runtime.strhash·f+0
```

&nbsp;

> 链接器将引用符号(`rel`) 数据(`1 = R_ADDR`) 填充到上面(十六进制数据) `偏移+长度` 位置。

&nbsp;

> 类型(`t`) 参考 `src/cmd/internal/objabi/reloctype.go`。</br>
> 数据段(`STEXT`, `SRODATA`) 参考 `src/cmd/internal/objabi/symkind.go`。

[Dave Cheney:Go’s hidden #pragmas](https://dave.cheney.net/2018/01/08/gos-hidden-pragmas)