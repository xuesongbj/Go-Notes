# make和new区别

## make
make编译器内置函数,用于分配并初始化一个类型对象(仅限于slice,map,chan类型)。

```
make([]int, 0, 1)
make(map[int]int, 1)
make(chan int)
```

根据创建类型不同,编译器会有不同的实现方式:

```
runtime.makemap			// map实现
runtime.makeslice		// slice实现
runtime.makechan 		// chan实现
```

## new
new函数被编译器编译之后由runtime.newobject函数实现。newobject函数直接调用mallocgc申请内存。

new函数返回指向申请类型的内存指针。

```
func newobject(typ *_type) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}
```

mallocgc 函数具体实现在"内存分配器"再详细阐述。