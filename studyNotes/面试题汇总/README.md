# 汇总

## 1. 实例1

```
func calc(index string, a, b int) int {
    ret := a + b
    fmt.Println(index, a, b, ret)
    return ret
}

func main() {
    a := 1
    b := 2
    defer calc("1", a, calc("10", a, b))
    a = 0
    defer calc("2", a, calc("20", a, b))
    b = 1
}
```

## 实例2:

```
func main() {
    s := make([]int, 5)
    s = append(s, 1, 2, 3)
    fmt.Println(s)
}
```

## 实例3: 以下代码能编译过去吗？为什么？

```
package main

import (
    "fmt"
)

type People interface {
    Speak(string) string
}

type Stduent struct{}

func (stu *Stduent) Speak(think string) (talk string) {
    if think == "bitch" {
        talk = "You are a good boy"
    } else {
        talk = "hi"
    }
    return
}

func main() {
    var peo People = Stduent{}
    think := "bitch"
    fmt.Println(peo.Speak(think))
}
```

## 实例4: 以下代码打印出来什么内容，说出为什么。。。

package main

import (
    "fmt"
)

type People interface {
    Show()
}

type Student struct{}

func (stu *Student) Show() {

}

func live() People {
    var stu *Student
    return stu
}

func main() {
    if live() == nil {
        fmt.Println("AAAAAAA")
    } else {
        fmt.Println("BBBBBBB")
    }
}

## 实例5: 下面代码能运行吗? 为什么?

```
type Param map[string]interface{}

type Show struct {
    Param
}

func main1() {
    s := new(Show)
    s.Param["RMB"] = 10000
}
```

## 实例6: 写出打印的结果

```
type People struct {
    name string `json:"name"`
}

func main() {
    js := `{
        "name":"11"
    }`
    var p People
    err := json.Unmarshal([]byte(js), &p)
    if err != nil {
        fmt.Println("err: ", err)
        return
    }
    fmt.Println("people: ", p)
}
```

## 实例7: 

```
// Stack overflow
// fmt.Sprintf 会有递归

type People struct {
    Name string
}

func (p *People) String() string {
    return fmt.Sprintf("print: %v", p)
}

func main() {
    p := &People{}
    p.String()
}
```

## 实例8: 

```
// 向一个关闭channel 发送数据
// panic: send on closed channel
func main() {
    ch := make(chan int, 1000)
    go func() {
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    go func() {
        for {
            a, ok := <-ch
            if !ok {
                fmt.Println("close")
                return
            }
            fmt.Println("a: ", a)
        }
    }()
    close(ch)
    fmt.Println("ok")
    time.Sleep(time.Second * 100)
}
```
