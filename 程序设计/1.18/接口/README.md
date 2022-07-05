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
