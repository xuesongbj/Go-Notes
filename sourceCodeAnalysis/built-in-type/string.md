# string
字符串类型是一个组合类型,由头部和底层数组两部分组成。头部由ptr、len组成。该ptr指向底层数组,len表示该数组长度。


## 源码剖析

### 数据类型
```
(gdb) ptype &a
type = struct string {
    uint8 *str;
    int len;
} *

```

### 类型转换
runtime/string.go


```
func slicebytetostring(buf *tmpBuf, b []byte) (str string) {
 	l := len(b)
 	if l == 0 {
   		return ""
 	}

 	// 如果字符串长度为1,则ptr指向该b[0]元素地址;len长度为1.
 	if l == 1 {
    	stringStructOf(&str).str = unsafe.Pointer(&staticbytes[b[0]])
   		stringStructOf(&str).len = 1
    	return
 	}

 	// buf == nil 表示有buf空间大小
   	// len([]byte) <= len(buf),如果是小对象,则直接放在当前G的buf空间。
   	// 没有了buf空间或者申请的空间大于32byte,则直接从heap上分配内存
   	// const tmpStringBufSize = 32
 	var p unsafe.Pointer
 	if buf != nil && len(b) <= len(buf) {
 		p = unsafe.Pointer(buf)
 	} else {
 		p = mallocgc(uintptr(len(b)), nil, false)
 	}

	// ptr指向底层数组首元素地址; len为该字节数组长度
 	stringStructOf(&str).str = p
 	stringStructOf(&str).len = len(b)

	// 最后进行内存拷贝操作
	// memmove具体实现详见: runtime/memmove_*.s
 	memmove(p, (*(*slice)(unsafe.Pointer(&b))).array, uintptr(len(b)))
 	return
}
```

### 无拷贝赋值
通常情况下将一个字符串变量内容赋值给另一个字符串变量时,会发生内存拷贝操作。为提升性能,go提供了无内存拷贝(两个变量同时指向同一块底层字节数组)。

runtime/string.go:
```
func gostringnocopy(str *byte) string {
	ss := stringStruct{str: unsafe.Pointer(str), len: findnull(str)}

	// 通过指针操作将s指向同一块儿底层字节数组
	s := *(*string)(unsafe.Pointer(&ss))
	return s
}
```

