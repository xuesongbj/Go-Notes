# sort

让目标对象实现特定接口，以支持排序。</br>
内部实现了 `QuickSort`、`HeapSort`、`InsertionSort`、`SymMerge`算法。

&nbsp;

```go
package main

import (
    "fmt"
    "sort"
)

func main() {
    s := []int{5, 2, 6, 3, 1, 4}
    sort.Ints(s)
    fmt.Println(s)
}
```

&nbsp;

## Slice

通过自定义函数，选择要比较的内容，或改变次序。

&nbsp;

```go
func Slice(x any, less func(i, j int) bool)
func SliceStable(x any, less func(i, j int) bool)

func SliceIsSorted(x any, less func(i, j int) bool)
```

&nbsp;

```go
package main

import (
    "fmt"
    "sort"
)

func main() {
    s := []struct{
        id   int
        name string
    }{
        {5, "a"},
        {2, "b"},
        {6, "c"},
        {3, "d"},
        {1, "e"},
        {4, "f"},
    }

    sort.Slice(s, func(i, j int) bool {
        return s[i].id > s[j].id
    })

    fmt.Println(s)
}

// [{6 c} {5 a} {4 f} {3 d} {2 b} {1 e}]
```

&nbsp;

## interface

避开辅助函数，实现排序接口。

```go
type Interface interface {
    // Len is the number of elements in the collection.
    Len() int

    // Less reports whether the element with index i 
    // must sort before the element with index j.
    Less(i, j int) bool

    // Swap swaps the elements with indexes i and j.
    Swap(i, j int)
}

func Sort(data Interface)       // 不稳定排序，不保证相等元素原始次序不变
func Stable(data Interface)     // 稳定排序，相等元素原始次序不变
```

```go
package main

import (
    "fmt"
    "sort"
)

type Data struct {
    text string
    index int
}

type Queue []Data

func (q Queue) Len() int {
    return len(q)
}

func (q Queue) Less(i, j int) bool {
    return q[i].index < q[j].index
}

func (q Queue) Swap(i, j int) {
    q[i], q[j] = q[j], q[i]
}


func main() {
    q := Queue{
        {"d", 3},
        {"c", 2},
        {"e", 4},
        {"a", 0},
        {"b", 1},
    }

    fmt.Println(sort.IsSorted(q))

    sort.Sort(q)
    fmt.Println(q, sort.IsSorted(q))
}

// false
// [{a 0} {b 1} {c 2} {d 3} {e 4}] true
```

&nbsp;

## Search

排序过后的数据，可用 `Search` 执行二分搜索(binary search)。</br>

返回 `[0,n)` 之间，`f() == true`的最小索引序号。 </br>

可用来查找有序插入位置。如找不到，则返回 `n`。

```go
func Search(n int, f func(int) bool) int
```

```go
package main

import (
    "fmt"
    "sort"
)

type Data struct {
    text string
    index int
}

type Queue []Data

func (q Queue) Len() int {
    return len(q)
}

func (q Queue) Less(i, j int) bool {
    return q[i].index < q[j].index
}

func (q Queue) Swap(i, j int) {
    q[i], q[j] = q[j], q[i]
}

func main() {
    q := Queue{
        {"d", 3},
        {"c", 2},
        {"e", 4},
        {"a", 0},
        {"b", 1},
    }

    sort.Sort(q)
    fmt.Println(q)

    i := sort.Search(len(q), func(index int) bool {
        return q[index].index > 6
    })

    fmt.Println("index > 6: ", i)

    i = sort.Search(len(q), func(index int) bool {
        return q[index].index >= 3
    })

    fmt.Println("index >= 3:", i)

    s := make(Queue, len(q) + 1)
    copy(s, q[:i])
    copy(s[i+1:], q[i:])
    s[i] = Data{"a3", 3}

    fmt.Println(s)
}

/*
[{a 0} {b 1} {c 2} {d 3} {e 4}]
index > 6:  5
index >= 3: 3
[{a 0} {b 1} {c 2} {a3 3} {d 3} {e 4}]
*/
```

&nbsp;

## Reverse

辅助函数 `Reverse` 返回一个将 `Less` 参数对调的包装对象。</br>
如此，判断结果就正好相反，实现倒序。

```go
// sort.go

type reverse struct {
    Interface
}

func (r reverse) Less(i, j int) bool {
    return r.Interface.Less(j, i)
}

func Reverse(data Interface) Interface {
    return &reverse{data}
}
```

```go
package main

import (
    "fmt"
    "sort"
)

type Data struct {
    text string
    index int
}

type Queue []Data

func (q Queue) Len() int {
    return len(q)
}

func (q Queue) Less(i, j int) bool {
    return q[i].index < q[j].index
}

func (q Queue) Swap(i, j int) {
    q[i], q[j] = q[j], q[i]
}



func main() {
    q := Queue{
        {"d", 3},
        {"c", 2},
        {"e", 4},
        {"a", 0},
        {"b", 1},
    }

    sort.Sort(sort.Reverse(q))
    fmt.Println(q)
}

// [{e 4} {d 3} {c 2} {b 1} {a 0}]
```