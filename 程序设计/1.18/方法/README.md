# 方法

方法是与对象实例相绑定的特殊函数。

* 方法是面向对象编程的基本概念，用于维护和展示对象自身状态。对象是内敛的，每个实例都有各自不同的独立特征，以属性和方法对外暴露。
* 普通函数专注于算法流程，接收参数完成逻辑运算，返回结果并清理现场。也就是说，方法有持续性状态，而函数通常没有。

1. 前置接收参数(receiver)，代表方法所属类型。</br>
2. 可为当前包，除接口和指针以外的任何类型定义方法。</br>
3. 不支持静态方法(static method)或关联函数。</br>
4. 不支持重载。</br>

&nbsp;

```go
func (int) test() {}
//    ~~~ cannot define new methods on non-local type int

type N *int
func (N) test() {}
//    ~ invalid receiver type N (pointer or interface type)

type M int
func (M) test(){}
func (*M) test(){}
//    ~~ redeclared in this block
```

&nbsp;

## 接收参数限制

对接收参数命名无限制，按管理选用简短有意义的名称。

&nbsp;

### 省略接收参数名

如方法内部不引用实例，可省略接收参数名，仅留类型。

```go
type N int

func (n N) toString() string {
    return fmt.Sprintf("%#x", n)
}

func (N) test() {      // 省略接收参数名。
    println("test")
}

// -----------------------------

func main() {
    var a N = 25
    println(a.toString())
}
```

&nbsp;

### 接收参数可以是指针类型

接收参数可以是指针类型，调用时据此判定是否复制(pass by value)。

> 注意区别: </br>
>   不能为指针和接口定义方法， 是说`N`本身不能是接口和指针。</br>
>   这与作为参数列表成员的 `receiver *N` 意思完全不同。 </br>
>   
>   </br>
>   方法本质上就是特殊的函数，接收参数无非是其第一参数。</br>
>   只不过，在某些语言里它是隐式的 `this`。

&nbsp;

```go
type N int

func (n N) copy() {
    fmt.Printf("%p, %v\n", &n, n)
}

func (n *N) ref() {
    fmt.Printf("%p, %v\n", n, *n)
}

// -----------------------------

func main() {
    var a N = 25
    fmt.Printf("%p\n", &a)  // 0xc000014080

    a.copy()                // 0xc000014088, 25
    N.copy(a)               // 0xc0000140a0, 25

    a++
    
    a.ref()                 // 0xc000014080, 26
    (*N).ref(&a)            // 0xc000014080, 26
}
```

编译器根据接收参数类型，自动在值和指针间转换。

&nbsp;

```go
type N int
func (n N) copy() {}
func (n *N) ref() {}

//---------------------

func main() {
    var a N = 25
    var p *N = &a

    a.copy()
    a.ref()     // (*N).ref(&a)

    p.copy()    // N.copy(*p)
    p.ref()
}
```

```go
$ go build -gcflags "-N -l"
$ go tool objdump -S -s "main\.main" ./test

TEXT main.main(SB)
func main() {
        var a N = 25
  0x455254              MOVQ $0x19, 0x8(SP)     
        var p *N = &a
  0x45525d              LEAQ 0x8(SP), CX        
  0x455262              MOVQ CX, 0x18(SP)       
        a.copy()
  0x455267              MOVQ 0x8(SP), AX        
  0x45526c              CALL main.N.copy(SB)    
        a.ref()
  0x455271              LEAQ 0x8(SP), AX        ; &a
  0x455276              CALL main.(*N).ref(SB)  
        p.copy()
  0x45527b              MOVQ 0x18(SP), CX       
  0x455282              MOVQ 0(CX), AX          
  0x45528a              CALL main.N.copy(SB)    
        p.ref()
  0x45528f              MOVQ 0x18(SP), AX       
  0x455294              CALL main.(*N).ref(SB)  
}
```

&nbsp;

不能以多级指针调用方法。

```go
func main() {
    var a N = 25
    var p *N = &a

    p2 := &p

    // p2.copy()
    // ~~~~~~~ p2.copy undefined

    // p2.ref()
    // ~~~~~~ p2.ref undefined

    (*p2).copy()
    (*p2).ref()
}
```

&nbsp;

### 如何确定接收参数(receiver)类型?

* 修改实例状态，用 `*T`。
* 不修改状态的小对象或固定值，用 `T`。
* 大对象用 `*T`，减少复制成本。
* 引用类型、字符串、函数等指针包装对象，用 `T`。
* 含 `Mutex` 等同步字段，用 `*T`，避免因复制造成锁无效。
* 其他无法确定的，都用 `*T`。