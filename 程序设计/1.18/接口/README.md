# 接口

接口(`interface`) 是多个方法声明集合，代表一种调用契约。

&nbsp;

在设计上，接口解除显式类型依赖(DIP，依赖倒置)，提供面向对象多态性。</br>
定义小型、灵活及组合性接口（ISP，接口隔离），减少可视方法，屏蔽内部结构和实现细节。</br>

&nbsp;

只要目标类型方法集包含接口全部方法，就视为实现该接口，无需显示声明。</br>
当然，单个目标类型可实现多个接口。

&nbsp;

* 接口限制
    * 不能有字段。
    * 只能生声明方法，不能实现。
    * 可嵌入其他接口。

* 接口实现
    * 通常以 `er` 作为名称后缀。
    * 空接口(`interface{}`，`any`) 没有任何方法声明。

&nbsp;

## 接口使用

### 注意方法集差异

```go
type Tester interface {
    Test()
    String(string) string  // 有效参数名便于阅读理解，不建议省略。
}

// -------------------------

type Data struct{}

func (*Data) Test() {}
func  (Data) String(s string) string { return "test:" + s }

// -------------------------

func main() {
    var d Data

    // var t Tester = d 
    //                ~ Data does not implement Tester 
    //                  Test method has pointer receiver

    var t Tester = &d
    t.Test()

    println(t.String("abc"))
}
```

&nbsp;

### 匿名接口

匿名接口可直接用于变量定义，或作为结构字段类型。

```go
type Data struct{}
func (Data) String() string { return "abc" }

// -------------------------

type Node struct {
    data interface {
        String() string
    }
}

// -------------------------

func main() {
    var t interface {
        String() string
    } = Data{}
    
    n := Node{
        data: t,
    }
    
    println(n.data.String())
}
```

&nbsp;

### 空接口

空接口可被赋值任何对象。

```go
func main() {
    var i interface{} = 123
    fmt.Println(i)

    i = "abc"
    fmt.Println(i)
}
```

&nbsp;

注意，接口会复制目标对象，通常以指针代替原始值。

```go
type Stringer interface {
    String() string
}

// -------------------------

type Data struct {
    X int
}

func (d Data) String() string { 
    return strconv.Itoa(d.X)
}

// -------------------------

func main() {
    d := Data{ 100 }
    var s Stringer = d   // copy

    d.X = 200

    println(d.String())  // 200
    println(s.String())  // 100
}
```

&nbsp;

## 空值判断

接口变量默认值 `nil`。如实现接口的类型支持，可做相等运算。

```go
func main() {
    var t1, t2 interface{}
    println(t1 == nil, t1 == t2)
    
    t1, t2 = 100, 100
    println(t1 == t2)
    
    // t1, t2 = map[string]int{}, map[string]int{}
    // println(t1 == t2)
    //        ~~~~~~~~ panic: comparing uncomparable type map
}
```

&nbsp;

接口内部由两个字段组成：**类型** 和 **值**。</br>

只有两个字段都为 `nil` 时，接口才等于 `nil`。可利用反射完善判断结果。

> 不能直接以指针判断 `data` 字段，因为编译器可能将其指向 `runtime.zerobase` 或 `zeroVal` 全局变量。

```go
// runtime/runtime2.go

type iface struct {
    tab  *itab
    data unsafe.Pointer
}

type eface struct {       // interface{}
    _type *_type
    data  unsafe.Pointer
}
```

```go
func main() {
    var d *Data
    var t Tester = d   // type != nil
    
    println(t == nil)  // false
    println(t == nil || reflect.ValueOf(t).IsNil()) // true
}
```

```go
import "reflect"

func main() {
    var t1 interface{}
    var t2 interface{} = ([]int)(nil)  // type != nil

    println(t1 == nil)  // true
    println(t2 == nil)  // false

    println(t1 == nil || reflect.ValueOf(t1).IsNil())  // true
    println(t2 == nil || reflect.ValueOf(t2).IsNil())  // true
}
```

&nbsp;

## 匿名嵌入

像匿名字段那样，嵌入其他接口。</br>

目标类型方法集中，必须全部方法实现，包括嵌入接口。</br>

* 嵌入相当于导入方法声明。(非继承)
* 不能嵌入自身或循环嵌入。
* 鼓励小接口嵌入组合。

&nbsp;

可以有相同签名(方法名、参数列表和返回值，不包括参数名)的方法声明。</br>
即便多个嵌入接口有相同声明亦是如此，因为最终方法集里相同声明仅有一个。

```go
type Aer interface {
    String(x string) string
}

// -------------------------

type Stringer interface {
    String(string) string
}

// -------------------------

type Tester interface {
    Aer
    Stringer                     // 多个嵌入接口有相同声明。

    Test()
    String(s string) string      // 签名相同（参数名不同）。

    // String() string           // 签名不同。
    // ~~~~~~ duplicate method
}
```

&nbsp;

## 类型转换

超集接口(即便非嵌入)可隐式转换为子集，反之不行。

```go
type Stringer interface {
    String(string) string
}

// -------------------------

type Tester interface {
    // Stringer

    Test()
    String(string) string
}

// -------------------------

type Data struct{}

func (*Data) Test() {}
func  (Data) String(s string) string { return "test:" + s }

// -------------------------

func main() {
    var t Tester = &Data{}

    var s Stringer = t
    s.String("abc")

    // var t2 Tester = s
    //                 ~ Stringer does not implement Tester
    
    // t2 := Tester(s)
    //              ~ Stringer does not implement Tester    
}
```

&nbsp;

以类型推断将接口还原为原始类型，或判断是否实现了某个更具体的接口类型。

```go
func main() {
    var t Tester = &Data{}

    var s Stringer = t
    s.String("abc")

    // 原始类型。
    d, ok := s.(*Data)
    fmt.Println(d, ok)          // &{} true

    // 其他接口。
    t2, ok := s.(Stringer)
    fmt.Println(t2, ok)         // &{} true

    // 和 fmt.Stringer 方法签名不同。
    // 如不用 ok-idiom，失败时 panic!
    t3, ok := s.(fmt.Stringer)
    fmt.Println(t3, ok)         // <nil> false
}
```

还可用 `switch` 语句在多种类型间做出推断匹配，如此空接口就有更多发挥空间。

* 未使用变量视为错误。
* 不支持 `fallthrough`。

```go
func main() {
    var i interface{} = &Data{}

    switch v := i.(type) {
    case nil: 
    case *int:
    case func()string:
    case *Data: fmt.Println(v)
    case Tester: fmt.Println(v)
    default:
    }
}
```

```go
func main() {
    var i interface{} = &Data{}

    switch v := i.(type) {   // v declared but not used
    case *Data: fallthrough  // fallthrough statement out of place
    default:
    }
}
```
