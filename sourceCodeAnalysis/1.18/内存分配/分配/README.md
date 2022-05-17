# 分配

按对象大小，选择不同分配策略。

```go
// malloc.go

// implementation of new builtin
func newobject(typ *_type) unsafe.Pointer {
    return mallocgc(typ.size, typ, true)
}
```

```go
// Allocate an object of size bytes.
// Small objects are allocated from the per-P cache's free lists.
// Large objects (> 32 kB) are allocated straight from the heap.

func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {

    // 零长度对象（zero）
    if size == 0 {
        return unsafe.Pointer(&zerobase)
    }
    
    if size <= maxSmallSize {
        if noscan && size < maxTinySize {
            // 微小对象（tiny）...
        } else {
            // 小对象（small）...
        }
    } else {
        // 大对象（large）...
    }

    return x
}
```
