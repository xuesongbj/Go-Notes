# stack

## 内存分配示意图

```go
                      +--------------------------+         +------------------+
           +--------> | mcache.stackcache[order] | ------> | stackcacherefill |
           |          +--------------------------+         +------------------+
           |                                                        |
         fixed                                                      |
           |                                                        |
           |                                                        v
     +------------+                                        +------------------+
---> | stackalloc |                                        | stackpoolalloc   |
     +------------+                                        +------------------+
           |                                               | stackpool[order] |
           |                 +-------------------+         +------------------+
         large        +----> | stackLarge.free[] |                  |
           |          |      +-------------------+                  |
           |          |                                             |
           +----------+                                             |
                      |                                             |
                      |      +-------------------+                  |
                      +----> | mheap.allocManual | <----------------+
                             +-------------------+
```

&nbsp;

## 内存释放示意图

```go
                      +--------------------------+        +-------------------+
           +--------> | mcache.stackcache[order] | -----> | stackcacherelease |
           |          +--------------------------+        +-------------------+
           |                                                        |
         fixed                                                      |
           |                                                        |
           |                                                        v
     +-----------+                                        +------------------+
---> | stackfree |                                        | stackpoolfree    |
     +-----------+                                        +------------------+
           |                                              | stackpool[order] |
           |                 +-------------------+        +------------------+
         large        +----> | stackLarge.free[] |                  |
           |          |      +-------------------+                  |
           |          |                                             |
           +----------+                                             |
                      |                                             |
                      |      +------------------+                   |
                      +----> | mheap.freeManual | <-----------------+
                             +------------------+
```
