# 泛型

泛型(generic)允许在强类型语言代码中使用**参数化类型**(parameterized types)。</br>
实例化时指明 **类型参数**(type parameter)， 有时也被称作模板。

* 函数和类型(含接口) 支持类型参数。
* 方法暂不支持，通过所属类型变通。(1.18)
* 支持推导，可省略类型实参(type argument)。

&nbsp;

```go
package main

import (
    "golang.org/x/exp/constraints"
)

func max[T constraints.Ordered](x, y T) T {
    if x > y { return x }
    return y
}

func main() {
    Println(max[int](1, 2))    // 实例化，类型实参。
    Println(max(1.1, 1.2))     // 类型推导。
}
```

```go
type Data[T any] struct {
    x T
}

func (d Data[T]) test() {
    fmt.Println(d)
}

// func (d Data[T]) test2[X any](x X) {}
//                       ~~~~~~~ method must have no type parameters

func main() {
    d := Data[int]{ x: 1 }
    d.test()
}
```

&nbsp;

鉴于 `interface{}` 很常用，专门为其引入别名 `any`。

> 其实 `[T any]` 直接简写为 `[T]` 更简洁一些。</br>
> 不知道后续版本是否会支持。

```go
type any = interface{}
```

&nbsp;

## 类型集合

接口的两种定义:

* **普通接口**：方法集合(method sets)。
* **类型约束**：类型集合(type sets)。

相比普通接口 「被动、隐式」实现，类型约束显式指定实现接口的类型集合。

```go
type Ordered interface {
    Integer | Float | ~string
}

type Integer interface {
    Signed | Unsigned
}

type Signed interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}
```

* 竖线(`|`) 表示类型集，匹配其中任一类型即可。
* 波浪线(`~`)表示底层类型(underlying type) 是该类型的所有类型。

* 内置: `comparable` (std)
* 附加: `golang.org/x/exp/constraints` (1.19+ merge)

```go
func main() {
    // println(max(struct{}{}, struct{}{}))
    //             ~~~~~~~~ struct{} does not implement Ordered
}
```

&nbsp;

普通接口可用作约束，但类型约束却不能当普通接口使用。

```go
type A int
type B string

func (a A) Test() { println("A:", a) }
func (a B) Test() { println("B:", b) }

// -----------------------------------

type Tester interface {
    A | B

    Test()
}

fun test[T Tester](x T) {
    x.Test()
}

// ------------------------------------

func main() {
    test[A](1)
    test(B)("abc")

    // var c Tester = A(1)
    //       ~~~~~~ interface contains type constraints
    // c.Test()
}
```

&nbsp;

## 约束类型

除含类型集合的接口类型外，也可直接写入参数列表。

```go
[T any]             // 任意类型
[T int]             // 只能是 int
[T ~int]            // 是 int 或底层类型是 int 的类型。(type I int)
[T int | string]    // 只能是 int 或 string。(interface{ int | string})
[T io.Reader]       // 任何实现io.Reader 接口的类型
```

```go
func test[T int | float32](x T) {
    fmt.Printf("%T, %v\n", x, x)
}

func main() {
    test(1)
    test[float32](1.1)

    // test("abc")
    //       ~~~ string does not implement int|float32
}
```

```go
func makeSlice[T int | float64](x T) []T {
    s := make([]T, 10)
    for i := 0; i < cap(s); i++ {
        s[i] = x
    }

    return s
}
```

```go
func test[T int | float64, E ~[]T](x E) {
    for v := range x {
        fmt.Println(v)
    }
}

func main() {
    test([]int{1, 2, 3})
    test([]float64{1.1, 1.2, 1.3})
}
```

* 泛型函数内不能定义新类型。
* 如类型约束不是接口，则无法调用其成员。

```go
type Num int
func (n Num) print() { println(n) }

func test[T Num](n T) {
    // type x int
    // ~~~~~~~~~~ type declarations inside generic functions are not currently supported

    // n.print()
    // ~~~~~~~ n.print undefined (type T has no field or method print)
}

func main() {
    test(1)
}
```

&nbsp;

## 类型转换

不支持switch 类型推断(type assert)，可先转换为普通接口。

```go
func test[T any](x T) {
    
    // switch x.(type) {}
    //          ~~~~ cannot use type switch on type parameter value x

    // 先转换为接口 ----------
    
    // var i interface{} = x
    var i any = x

    switch i.(type) {
    case int: println("int", x)
    }
}
```

[An Introduction To Generics](https://go.dev/blog/intro-generics)