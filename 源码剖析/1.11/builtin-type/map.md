# Map

## 哈希表实现
结构上用一个头结构（hmap）存储基本信息。 最重要的是引用 buckets 数组。这些桶（bmap）存储具体 key/value 数据。

```
+--------------+
| hmap         |
+--------------+
|   buckets    |   header
+--------------+
       |
       |
       v
+--------------+---//---+------+-------------+
| bmap         |   ...  | bmap | prealloc... |   bucket array
+--------------+---//---+------+-------------+
|   [8]tophash |
+--------------+
|   [8]key     |
+--------------+
|   [8]value   |
+--------------+
|   overflow   |
+--------------+
       |
       |
       v
+--------------+
| bmap         |
+--------------+ 
```

### 头结构
Maximum average load of a bucket that triggers growth is 6.5.

```
// A header for a Go map.
type hmap struct {
    count      int             // 键值对数量（len）。
    flags      uint8           // 状态标记。
    B          uint8           // 可容纳的键值对数量 = loadFactor * 2^B
    noverflow  uint16          // 溢出桶数量（创建溢出桶时累加）。
    hash0      uint32          // 哈希种子。

    buckets    unsafe.Pointer  // 桶数组地址。
    oldbuckets unsafe.Pointer  // 旧桶数组地址，用于扩容。
    nevacuate  uintptr         // 迁移进度，低于此值的已安置。

    extra      *mapextra       // optional fields
}
```

### 溢出桶

```
type mapextra struct {
 // overflow和oldoverflow存储指向所有溢出桶的指针
 // overflow和oldoverflow仅在key和value不包含指针时使用.
 // overflow包含hmap.buckets溢出桶
 // oldoverflow包含hmap.oldbuckets溢出桶
 overflow    *[]*bmap
 oldoverflow *[]*bmap

 nextOverflow   *bmap       // 预分配桶。
}
```

### 桶结构

* 每个桶内存储 8 个键值对。
* 如果key和value数据长度不能超过128byte,直接在桶内存储.否则会在堆分配内存,间接存储。
* 数组tophash内存储key哈希高位值,用于访问时快速匹配.只有tophash值相等,才去匹配具体的key。
* 根据键值长度,动态分配键值对存储分配。

```
// A bucket for a Go map.
type bmap struct {

    // tophash generally contains the top byte of the hash value
    // for each key in this bucket. If tophash[0] < minTopHash,
    // tophash[0] is a bucket evacuation state instead.

    tophash [bucketCnt]uint8

    // Followed by bucketCnt keys and then bucketCnt values.
    // Followed by an overflow pointer.
} 
```

![map_struct](./images/map_1.jpg)


## 创建map

### 程序入口
首先找到创建map入口函数runtime.makemap。

```
0x000000000104b510 <+64>: mov    QWORD PTR [rsp],rax
0x000000000104b514 <+68>: mov    QWORD PTR [rsp+0x8],0xa
0x000000000104b51d <+77>: lea    rax,[rsp+0x38]
0x000000000104b522 <+82>: mov    QWORD PTR [rsp+0x10],rax
0x000000000104b527 <+87>: call   0x1005f40 <runtime.makemap>
0x000000000104b52c <+92>: mov    rax,QWORD PTR [rsp+0x18]
```

### 具体实现

主要是计算出合理大小，并创建桶数组。

* 如果 make(..., 100)，那么计算出来的 B = 4，可容纳 6.5 * 2^4 = 104 个键值对。 以 map\[string\]int 为例，那么每个桶大小是 tophash_8 + key_string_16 * 8 + value_int_8 * 8 + overflow_8 = 208。 使用 make(..., hint) 预分配空间，有助于减少扩容，提升性能。

```
 // makemap implements Go map creation for make(map[k]v, hint).
// If the compiler has determined that the map or the first bucket
// can be created on the stack, h and/or bucket may be non-nil.
// If h != nil, the map can be created directly in h.
// If h.buckets != nil, bucket pointed to can be used as the first bucket.
func makemap(t *maptype, hint int, h *hmap) *hmap {
 // 创建头,初始化
 if h == nil {
     h = new(hmap)
 }
 // hash种子
 h.hash0 = fastrand()

 // 根据hint,计算B指数
 // 根据该算法,如果hint==100,则B指数为4.
 B := uint8(0)
 for overLoadFactor(hint, B) {
     B++
 }
 h.B = B

 // 分配并初始化哈希桶
 // 如果B == 0,则buckets延迟分配。
 if h.B != 0 {
     var nextOverflow *bmap

     // 分配内存,创建桶数组
     h.buckets, nextOverflow = makeBucketArray(t, h.B, nil)

     // 为减少内存分配操作, 分配的桶可能超过预期
     // 将其存放起来，做为后面溢出桶的预分配
     if nextOverflow != nil {
         h.extra = new(mapextra)
         h.extra.nextOverflow = nextOverflow
     }
 }

 return h
}

```

### 指数计算公式
```
// 计算B是否大于loadFactor
// 如果hint==100,则B==4,可容纳6.5 * 2^4 = 104 个键值对。
func overLoadFactor(count int, B uint8) bool {
 	return count > bucketCnt && uintptr(count) >loadFactorNum*(bucketShift(B)/loadFactorDen)
}
```

### 创建哈希桶

* 为哈希桶创建和初始化底层数组。
* 1<<b是创建的最小哈希桶数量。
* 如果dirtyalloc==nil,则创建新的哈希桶;否则原来哈希桶被清理,然后重新被底层数组使用。

```
// makeBucketArray initializes a backing array for map buckets.
// 1<<b is the minimum number of buckets to allocate.
// dirtyalloc should either be nil or a bucket array previously
// allocated by makeBucketArray with the same t and b parameters.
// If dirtyalloc is nil a new backing array will be alloced and
// otherwise dirtyalloc will be cleared and reused as backing array.
func makeBucketArray(t *maptype, b uint8, dirtyalloc unsafe.Pointer) (buckets unsafe.Pointer, nextOverflow *bmap) {
 base := bucketShift(b)
 nbuckets := base

 // 如果指数大于4,则预分配溢出桶
 if b >= 4 {
     nbuckets += bucketShift(b - 4)
     sz := t.bucket.size * nbuckets
     up := roundupsize(sz)
     if up != sz {
         nbuckets = up / t.bucket.size
     }
 }

 if dirtyalloc == nil {
     // 创建新的哈希桶
     buckets = newarray(t.bucket, int(nbuckets))
 } else {
     // 如果之前存在哈希桶,之前的哈希桶被清理.然后,重新被底层数组使用
     buckets = dirtyalloc
     size := t.bucket.size * nbuckets
     if t.bucket.kind&kindNoPointers == 0 {
         memclrHasPointers(buckets, size)
     } else {
         memclrNoHeapPointers(buckets, size)
     }
 }

 if base != nbuckets {
     // 提前预分配哈希桶
     nextOverflow = (*bmap)(add(buckets, base*uintptr(t.bucketsize)))
     last := (*bmap)(add(buckets, (nbuckets-1)*uintptr(t.bucketsize)))
     last.setoverflow(t, (*bmap)(buckets))
 }
 return buckets, nextOverflow
} 
```

## 读取
计算查找 key 的哈希值,以此找到对应的桶(bmap,bucket)。

* 先用哈希高位值在bmap.tophash数组里匹配.（整数要比直接匹配 key 高效）
* 如果 tophash 匹配成功，需要进一步检查 key 是否相等。（有 tophash 过滤，就不用每次匹配 key）
* 根据索引计算 value 地址，并返回。

```
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {

 // 计算key哈希值,根据公式找到对应的桶
 // hahs & m, 判断hash在哪个桶内,结果一定小于m
 alg := t.key.alg        // 算法表
 hash := alg.hash(key, uintptr(h.hash0))
 m := bucketMask(h.B)
 b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))   // buckets ptr + buckets offset(bucketssize), 找到所在的桶


 // 扩容迁移是分多次完成的。因此会有某些数据暂时留在oldbuckets,先从此查找。
 // 如果处于扩容迁移状态时, old bucket + overflow >= buckets,所以直接从old buckets+overflow肯定可以查询到key的值,否则该key不存在于该map.
 if c := h.oldbuckets; c != nil {
     // 扩容大小不是一个新的map大小
     if !h.sameSizeGrow() {
         // There used to be half as many buckets; mask down one more power of two.
         m >>= 1
     }
     oldb := (*bmap)(add(c, (hash&m)*uintptr(t.bucketsize)))

     // 检查是否已经完成迁移
     if !evacuated(oldb) {
         // 没有迁移完成,或还没有开始迁移
         // 则从old buckets查找key
         b = oldb
     }
 }

 // 计算key哈希高位值
 top := tophash(hash)

 // 遍历该桶,以及溢出链表。
 // 方法overflow通过bmap.overflow存储的地址找到链表下一个溢出桶。
 for ; b != nil; b = b.overflow(t) {

     // 遍历桶内的 tophash 数组，初步匹配。
     for i := uintptr(0); i < bucketCnt; i++ {

         // 不等,检查下一个
         if b.tophash[i] != top {
             continue
         }

         // 取出存储的key,判断是否相等
         k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
         if t.indirectkey {
             k = *((*unsafe.Pointer)(k))
         }
         if alg.equal(key, k) {
             // 返回对应value地址
             v := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
             if t.indirectvalue {
                 v = *((*unsafe.Pointer)(v))
             }
             return v
         }
     }
 }
 return unsafe.Pointer(&zeroVal[0])
}
```


![map_read](./images/map_2.jpg)



## 写入

区分插入和更新操作两种。

* 如果是更新操作，只需遍历找到并返回 value 地址即可。
* 如果桶有空位，则存入 tophash 和 key，然后返回 value 地址。
* 如果没有空位，新建溢出桶。然后使用该桶第一个空位。

```
func mapassign(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
 // 不能对空map进行写入
 if h == nil {
     panic(plainError("assignment to entry in nil map"))
 }

 // 当前已经有goroutine对map进行写操作
 if h.flags&hashWriting != 0 {
     throw("concurrent map writes")
 }

 // 计算key hash值
 alg := t.key.alg
 hash := alg.hash(key, uintptr(h.hash0))

 // 如果bucket为空,则创建bucket
 if h.buckets == nil {
     h.buckets = newobject(t.bucket) // newarray(t.bucket, 1)
 }

again:
 // 计算对应的桶索引
 bucket := hash & bucketMask(h.B)

 // 扩容,检查该桶是否需要做迁移
 if h.growing() {
     growWork(t, h, bucket)
 }

 // 找到key对应bucket
 b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))

 // 计算哈希高位值
 top := tophash(hash)

 var inserti *uint8
 var insertk unsafe.Pointer
 var val unsafe.Pointer

 // 1. 如果桶有空位,则存储tophash和key,然后返回value地址。
 // 2. 如果是更新操作,只需遍历找到并返回value地址即可。
 for {
     // 桶内匹配,查找存放位置
     for i := uintptr(0); i < bucketCnt; i++ {
         // 1. 如果和tophash不等
         // 2. 当前位置为空,计算索引、key/value存储地址
         // 3. inserti赋值,决定了只找第一个插入位置
         if b.tophash[i] != top {
             if b.tophash[i] == empty && inserti == nil {
                 inserti = &b.tophash[i]
                 insertk = add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
                 val = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
             }
             continue
         }

         // 如果b.tophash[i] == top,继续匹配key,可能是更新操作.
         // 返回value地址,供写入新值
         k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
         if t.indirectkey {
             k = *((*unsafe.Pointer)(k))
         }

         // 类型判断
         if !alg.equal(key, k) {
             continue
         }

         // already have a mapping for key. Update it.
         if t.needkeyupdate {
             typedmemmove(t.key, k, key)
         }

         // 插入更新值地址
         val = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))

         // 更新操作无需更多处理
         goto done
     }

     // 如果当前桶没有匹配,则继续检查溢出桶
     ovf := b.overflow(t)
     if ovf == nil {
         break
     }
     b = ovf
 }

 // 如果负载过大,或者有太多溢出桶,则进行扩容.
 if !h.growing() && (overLoadFactor(h.count+1, h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
     hashGrow(t, h)
     goto again // Growing the table invalidates everything, so try again
 }

 // 如果没找到插入或者更新位置,则意味着需要建立新的溢出桶
 // 优先使用预分配的溢出桶
 if inserti == nil {
     // all current buckets are full, allocate a new one.
     newb := h.newoverflow(t, b)
     inserti = &newb.tophash[0]
     insertk = add(unsafe.Pointer(newb), dataOffset)
     val = add(insertk, bucketCnt*uintptr(t.keysize))
 }

 // 存储key/value。(间接需要额外分配内存,存储指针)
 if t.indirectkey {
     kmem := newobject(t.key)
     *(*unsafe.Pointer)(insertk) = kmem
     insertk = kmem
 }
 if t.indirectvalue {
     vmem := newobject(t.elem)
     *(*unsafe.Pointer)(val) = vmem
 }
 typedmemmove(t.key, insertk, key)

 // 更新top hash和map键/值对数量
 *inserti = top
 h.count++

done:
 // 判断当前是否有其它线程对该map进行写入操作
 if h.flags&hashWriting == 0 {
     throw("concurrent map writes")
 }

 // 获取map写入控制权
 h.flags &^= hashWriting
 if t.indirectvalue {
     val = *((*unsafe.Pointer)(val))
 }

 // 返回value地址
 return val
}
```


## 删除
删除操作不会收缩内存，只是清理内容。

```
func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
 // 如果h是一个空map或者h键/值对为空,直接退出
 if h == nil || h.count == 0 {
     return
 }

 // 计算key哈希值,以及找到对应的桶索引
 alg := t.key.alg
 hash := alg.hash(key, uintptr(h.hash0))
 bucket := hash & bucketMask(h.B)

 // 判断当前是否在进行迁移操作
 if h.growing() {
     growWork(t, h, bucket)
 }

 // 找到对应的桶
 // 计算hash高位值
 b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
 top := tophash(hash)
search:

 // 遍历桶链表
 // 如果桶链表没有找到,则遍历overflow
 for ; b != nil; b = b.overflow(t) {
     // 桶内遍历
     for i := uintptr(0); i < bucketCnt; i++ {

         // tophash不等,跳过
         if b.tophash[i] != top {
             continue
         }

         // 匹配key,不等就跳过
         k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
         k2 := k
         if t.indirectkey {
             k2 = *((*unsafe.Pointer)(k2))
         }
         if !alg.equal(key, k2) {
             continue
         }

         // 相等,清理key内容
         if t.indirectkey {
             *(*unsafe.Pointer)(k) = nil
         } else if t.key.kind&kindNoPointers == 0 {
             memclrHasPointers(k, t.key.size)
         }

         // 清理value内容
         v := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
         if t.indirectvalue {
             *(*unsafe.Pointer)(v) = nil
         } else if t.elem.kind&kindNoPointers == 0 {
             memclrHasPointers(v, t.elem.size)
         } else {
             memclrNoHeapPointers(v, t.elem.size)
         }

         // 清理tophash内容,减少计数
         b.tophash[i] = empty
         h.count--

         // 跳出循环
         break search
     }
 }
}
```

### 编译器对遍历删除map元素实现
清空（clear）也只是重置内存，不会收缩. 当range delete时,编译器会将其优化成清空操作。
```
func mapclear(t *maptype, h *hmap) {
 // 如果空map或map键/值对为空
 // 直接退出
 if h == nil || h.count == 0 {
     return
 }


 // 重置属性
 h.flags &^= sameSizeGrow
 h.oldbuckets = nil
 h.nevacuate = 0
 h.noverflow = 0
 h.count = 0

 if h.extra != nil {
     *h.extra = mapextra{}
 }

 // 清理桶内数据
 _, nextOverflow := makeBucketArray(t, h.B, h.buckets)
 if nextOverflow != nil {
     // If overflow buckets are created then h.extra
     // will have been allocated during initial bucket creation.
     h.extra.nextOverflow = nextOverflow
 }
}
```

## 扩容
在写操作（mapassign）里会判断是否需要扩容。

* 扩容基本上是2x容量

扩容步骤:

* 检查是否达到负载系数.
* 超过,则增加B指数,横向扩容桶数组.
* 将现有桶转移到oldbuckets.
* 如果需要提前预分配,创建溢出桶.

### buckets扩容

```
// growing reports whether h is growing. The growth may be to the same size or bigger.
func (h *hmap) growing() bool {
	return h.oldbuckets != nil
} 
```

```
func hashGrow(t *maptype, h *hmap) {
 // 检查是否达到负载系数
 // 超过,则增加B指数,横向扩容桶数组
 bigger := uint8(1)
 if !overLoadFactor(h.count+1, h.B) {
     bigger = 0
     h.flags |= sameSizeGrow
 }

 // 将现有桶数组转移到oldbuckets
 // 创建新数组
 oldbuckets := h.buckets
 newbuckets, nextOverflow := makeBucketArray(t, h.B+bigger, nil)

 // commit the grow (atomic wrt gc)
 h.B += bigger           // 指数增加
 h.flags = flags
 h.oldbuckets = oldbuckets   // 将现buckets数据转移到oldbucets
 h.buckets = newbuckets      // 创建新buckets
 h.nevacuate = 0         // 迁移进度
 h.noverflow = 0         // 溢出桶数量


 // 如果需要提前预分配
 // 则提前预分配操作
 if nextOverflow != nil {
     if h.extra == nil {
         h.extra = new(mapextra)
     }
     h.extra.nextOverflow = nextOverflow
 }

 // the actual copying of the hash table data is done incrementally
 // by growWork() and evacuate().
}
```

![map_grow](./images/map_3.jpg)



### oldbucket迁移

迁移分2次:

* 进行写入/删除时,随机对bucket进行迁移。
* 按照h.nevacuate顺序进行迁移。

```
func growWork(t *maptype, h *hmap, bucket uintptr) {
 evacuate(t, h, bucket&h.oldbucketmask())

 if h.growing() {
     evacuate(t, h, h.nevacuate)
 }
} 
```

* 找到当前迁移桶位置(随机).
* 计算oldbucket桶长度.
* 如果桶还没有进行迁移,则进行迁移
   * 新桶分为前后两部分
   * 遍历旧桶和溢出桶
   * 对迁移数据key重新哈希,确定新的位置.然后将top、key、value复制到新桶
* 如果进行顺序迁移,对迁移进度计数器进行处理(用来判断是否迁移完成)


```
func evacuate(t *maptype, h *hmap, oldbucket uintptr) {
 // 迁移桶位置
 b := (*bmap)(add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
 // 计算桶增长之前的数量
 newbit := h.noldbuckets()

 // 如果该桶还没有迁移
 if !evacuated(b) {
     // 将新桶分成前后两部分
     var xy [2]evacDst

     // 新桶前部分
     // x.b: 新桶前半段地址
     x := &xy[0]
     x.b = (*bmap)(add(h.buckets, oldbucket*uintptr(t.bucketsize)))
     x.k = add(unsafe.Pointer(x.b), dataOffset)
     x.v = add(x.k, bucketCnt*uintptr(t.keysize))

     if !h.sameSizeGrow() {

         // 新桶后半段地址
         y := &xy[1]
         y.b = (*bmap)(add(h.buckets, (oldbucket+newbit)*uintptr(t.bucketsize)))
         y.k = add(unsafe.Pointer(y.b), dataOffset)
         y.v = add(y.k, bucketCnt*uintptr(t.keysize))
     }


     // 遍历旧桶及溢出桶
     for ; b != nil; b = b.overflow(t) {

         // key, value 地址
         k := add(unsafe.Pointer(b), dataOffset)
         v := add(k, bucketCnt*uintptr(t.keysize))

         // 桶内遍历
         for i := 0; i < bucketCnt; i, k, v = i+1, add(k, uintptr(t.keysize)), add(v, uintptr(t.valuesize)) {
             top := b.tophash[i]
             if top == empty {
                 b.tophash[i] = evacuatedEmpty
                 continue
             }

             // 重新hash(key)确定新位置(X or Y)…
             // 将top、key、value复制到新桶…
         }
     }
 }

 // 顺序迁移时,参数oldbucket就是顺序序号,所以相等.
 // 如此,增加顺序号,找到下一个要处理的桶,做为下次迁移目标.
 // 当然,随机时也可能相同,那就表示该桶已被处理,同样增加顺序号.
 if oldbucket == h.nevacuate {
     advanceEvacuationMark(h, t, newbit)
 }
}
```

![map_grow](./images/map_4.jpg)


* 估计n.nevacuate 边界
* 检查下一个桶是否迁移完成,如果迁移完成,n.nevacuate序号加1
* 检查所有旧桶是否迁移完成,如果迁移完成。释放旧桶

```
func advanceEvacuationMark(h *hmap, t *maptype, newbit uintptr) {
 h.nevacuate++

 // h.nevacuate 迁移进度计数器,每次最多加1024
 // 估计边界
 stop := h.nevacuate + 1024
 if stop > newbit {
     stop = newbit
 }

 // 检查下一个桶是否已经迁移完成。
 // 确定实际顺序序号(如果超出,则表示全部迁移完成)
 for h.nevacuate != stop && bucketEvacuated(t, h, h.nevacuate) {
     h.nevacuate++
 }

 // 如果所有旧桶迁移完毕,释放旧桶.
 // 不会再引发迁移操作
 if h.nevacuate == newbit { // newbit == # of oldbuckets
     h.oldbuckets = nil

     if h.extra != nil {
         h.extra.oldoverflow = nil
     }
     h.flags &^= sameSizeGrow
 }
}
```

## 其它
indirectkey 和 indirectvalue存储的是类型本身还是类型指针? 取决于类型大小,当大于128字节时。indirectkey和indirectkey会使用指针。当使用指针时,GC进行扫描时,会更加耗时.


```
// Maximum key or value size to keep inline (instead of mallocing per element).
// Must fit in a uint8.
// Fast versions cannot handle big values - the cutoff size for
// fast versions in cmd/compile/internal/gc/walk.go must be at most this value.
 maxKeySize   = 128
 maxValueSize = 128
```

### 不使用indirectkey和indirectvalue

```
package main

import "fmt"

func main() {
    type P struct {
        Age [16]int         // 128 Bytes
    }
    var a = map[P]P{}
    a[P{}] = P{}
    fmt.Println(a)
}
```

* gdb 调试

```
(gdb) ptype *t
type = struct runtime.maptype {
    runtime._type typ;
    runtime._type *key;
    runtime._type *elem;
    runtime._type *bucket;
    runtime._type *hmap;
    uint8 keysize;
    bool indirectkey;
    uint8 valuesize;
    bool indirectvalue;
    uint16 bucketsize;
    bool reflexivekey;
    bool needkeyupdate;
}
(gdb) p/x t.keysize             => 0x80 => [16]int => 16*8 = 128
$1 = 0x80
(gdb) p/x t.valuesize
$2 = 0x80
(gdb) p/x t.indirectkey          => 0x0 false
$3 = 0x0
(gdb) p/x t.indirectvalue
$4 = 0x0
```
通过调试信息可以看到,key和value小于128Bytes,会存储类型本身。


### 使用indirectkey和indirectvalue

```
package main

import "fmt"

func main() {
    type P struct {
        Age [16]int
        David bool
    }
    
    var a = map[P]P{}
    a[P{}] = P{}
    fmt.Println(a)
}
```

* gdb 调试

```
(gdb) ptype *t
type = struct runtime.maptype {
    runtime._type typ;
    runtime._type *key;
    runtime._type *elem;
    runtime._type *bucket;
    runtime._type *hmap;
    uint8 keysize;
    bool indirectkey;
    uint8 valuesize;
    bool indirectvalue;
    uint16 bucketsize;
    bool reflexivekey;
    bool needkeyupdate;
}
(gdb) p/x t.keysize
$1 = 0x8
(gdb) p/x t.valuesize
$2 = 0x8
(gdb) p/x t.indirectkey
$3 = 0x1
(gdb) p/x t.indirectvalue
$4 = 0x1
```
当P结构体增加一个bool类型后, 再次查看indirectkey和indirectvalue均采用指针存储。