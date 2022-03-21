# generic(泛型)

为了支持泛型函数，Go1.18 开始支持泛型参数，泛型参数的操作必须是所有参数类型能够支持的。

## 约束(constraint)

仅允许使用该约束范围内进行操作泛型。目前Go支持以下泛型约束：

* `any`：无约束
* `comparable`：约定范围是泛型能够使用`==`和`!=`进行比较的类型
* `custom`：自定义约束，Eg: `V int 64 | float64`，表示`V`类型可以是`int64`或`float64`

&nbsp;

### any

`any` 其实是 `interface{}` 的别名，行为和`interface{}`相同，让人看起来跟舒服，以后可以不再使用`interface{}`。

```go
// ForEach使用any配合类型参数，可以让ForEach()接收任何类型的参数

package main

import (
    "fmt"
)

func ForEach[T any](list []T, action func(T)) {
    for _, item := range list {
        action(item)
    }
}

func main() {
    ForEach([]string{"Hello", "world"}, func(s string) {
        fmt.Println(s)
    })
}
```

如果不使用类型参数，可以使用强制类型转换：

```go
// any类型的参数强制转换成string类型
package main

import (
    "fmt"
)

func ForEachWithAny(list []any, action func(any)) {
   for _, item := range list {
      action(item)
   }
}

func main() {
    ForEachWithAny([]any{"Hello", "World"}, func(s any) {
        fmt.Printf(s.(string))
    })
}
```

&nbsp;

### comparable

`comparable`属于go预先声明的(`buildin.go`)，它表示任何能够使用`==`和`!=`进行比较的类型。使用该约束条件的泛型必须支持该操作，否则报错。

```go
package main

import (
    "fmt"
)

func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
    var s V
    
    for _, v := range m {
        s += v
    }
    return s
}

func main() {
    ints := map[string]int64{
        "first":  34,
        "second": 12,
    }
    
    floats := map[string]float64{
        "first":  35.98,
        "second": 26.99,
    }
    
    fmt.Printf("sum: %v and %v\n", SumIntsOrFloats[string, int64](ints), SumIntsOrFloats[string, float64](floats))

    // 可忽略类型参数，编译器可以自动推断
    // fmt.Printf("sum: %v and %v\n", SumIntsOrFloats(ints), SumIntsOrFloats(floats))
}
```

&nbsp;

### custom

自定义约束，可以根据实际需要进行定义约束条件。

```go
package main

import (
    "fmt"
)

type Number interface {
    int64 | float64
}

func SumNumbers[K comparable, V Number](m map[K]V) V {
    var s V
    for _, v := range m {
        s += v
    }
    return s
}

func main() {
    ints := map[string]int64{
        "first":  34,
        "second": 12,
    }
    
    floats := map[string]float64{
        "first":  35.98,
        "second": 26.99,
    }
    
    fmt.Printf("sum: %v and %v\n", SumNumbers[string, int64](ints), SumNumbers[string, float64](floats))
}
```

&nbsp;

#### 衍生类型(`~`)

如下实例，类型`ID`和类型`int64`是不同的，因此无法使用上面定义的`SumNumbers()`。Go1.18提供了操作符`~`表示衍生类型，这样可以使用`~int64`表示`int64`和它任何衍生类型。

```go
package main

import (
    "fmt"
)

type ID int64

type NumberDerived interface {
   ~int64 | ~float64
}

func SumNumbersDerived[K comparable, V NumberDerived](m map[K]V) V {
   var s V
   for _, v := range m {
      s += v
   }
   return s
}


func main() {
    ids := map[string]ID{
        "first":  ID(34),
        "second": ID(12),
    }
    fmt.Printf("sum: %v\n", SumNumbersDerived(ids))
}
```

&nbsp;

## 源码剖析

### 实例代码

```go
package main

import (
    "errors"
    "fmt"
)

func indexOf[T comparable](s []T, x T) (int, error) {
    for i, v := range s {
        if v == x {
            return i, nil
        }
    }
    return 0, errors.New("not found")
}

func main() {
    var s []string = []string{"apple", "banana", "pear"}
    i, err := indexOf(s, "banana")
    fmt.Println(i, err)

    var it []int = []int{1, 2, 3}
    n, err := indexOf(it, 1)
    fmt.Println(n, err)
}
```

### Go 泛型实现

`generic` 在很多静态编译型语言都存在该能力。`generic` 发生在编译期，对`Runtime`性能没有影响。Go 1.18中通过字典和`GCShape`模板的泛型实现。在模板法(`GCShape`)和字典法(`.dict`)中间取一个中庸。

&nbsp;

#### GCShape

`GCShape`是一组类型，实例化时指定类型参数就可以共享通用函数(方法)签名。`GCShape`类型是我们在实例化实现中使用的特定类型，用于填充`gcshape`分组的所有类型。

&nbsp;

#### 字典(`.dict`)

泛型的编译器实现主要关注创建泛型函数和方法的实例化，这些函数和方法将使用具有具体类型的参数执行。

为了避免调用泛型函数(方法)时，根据不同函数参数生成不同的函数签名。并且为了实现简单(和性能)，没必要每次调用泛型函数时生成一个函数签名，我们可以在具有相同`GCShape`的类型参数之间共享泛型函数的签名。

&nbsp;

#### 模板(`go.shape.{type}_{number}`)

使用泛型编程，编译器编译时根据实际类型对函数泛型重命名(`go.shape.string_0`)。使用模板方法的优势是，所有工作都在编译器完成，对运行时(`runtime`)没有性能影响。但带来的问题是，如果代码中存在大量函数(方法)调用，则会产生大量的函数签名，使编译后的ELF文件`.TEXT`过大。

针对`.text`过大问题，所以Go并未完全采用模板法，而是采用模板法和字典法的一种折中方案。

&nbsp;

#### 代码片段

```go
func main() {
    var s []string = []string{"apple", "banana", "pear"}
    i, err := indexOf(s, "banana")
    fmt.Println(i, err)
}
```

```x86asm
; 被调函数indexOf，根据实参类型，将泛型进行替换为实际类型。
; 同时修改indexOf函数签名为 indexOf[go.shape.string_0] 
=> 0x000000000047e150 <+240>:	movups XMMWORD PTR [rsp+0xb0],xmm15
   0x000000000047e159 <+249>:	mov    rbx,QWORD PTR [rsp+0xe0]
   0x000000000047e161 <+257>:	mov    rcx,QWORD PTR [rsp+0xe8]
   0x000000000047e169 <+265>:	mov    rdi,QWORD PTR [rsp+0xf0]
   0x000000000047e171 <+273>:	lea    rax,[rip+0x34a88]        # 0x4b2c00 <main..dict.indexOf[string]>
   0x000000000047e178 <+280>:	lea    rsi,[rip+0x16ddd]        # 0x494f5c
   0x000000000047e17f <+287>:	mov    r8d,0x6
   0x000000000047e185 <+293>:	call   0x47e4e0 <main.indexOf[go.shape.string_0]>

; 变量s类型 --> slice --> []string
(gdb) ptype s
type = struct []string {
    string *array;
    int len;
    int cap;
}

; string结构 --> |ptr|len| --> 底层数组
; bx寄存器存储 字符串指针
(gdb) x/6xg $rbx
0xc000064f40:	0x0000000000494d87	0x0000000000000005      ; s[0]  -> apple
0xc000064f50:	0x0000000000494f5c	0x0000000000000006      ; s[1]  -> banana
0xc000064f60:	0x0000000000494cc7	0x0000000000000004      ; s[2]  -> pear

; apple -> 5byte
(gdb) x/5xb 0x0000000000494d87
0x494d87:	0x61	0x70	0x70	0x6c	0x65

; banana -> 6byte
(gdb) x/6xb 0x0000000000494f5c
0x494f5c:	0x62	0x61	0x6e	0x61	0x6e	0x61

; pear -> 4byte
(gdb) x/4xb 0xc000064f60
0xc000064f60:	0xc7	0x4c	0x49	0x00

; 1. go.shape.string_0 -> 根据参数生成的函数签名模板，如果再有相同的请求
; 2. .dict -> 通过查询.dict字典，获取函数参数类型
; 3. s, x 为参数列表
; 4. ~r0, ~r1 为函数返回值，因为Go1.17后，参数和返回值通过寄存器进行数据传递
main.indexOf[go.shape.string_0] (.dict=4926464, s=..., x=..., ~r0=10, ~r1=...) at /root/go/david/main.go:8
```
