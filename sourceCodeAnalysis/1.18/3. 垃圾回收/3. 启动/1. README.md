# 启动

```go
                                         +--------------------------+
                                         | gcStart                  |
                                         +--------------------------+
+------------+                           |  for { sweepone }        |
| mallocgc   | --- TriggerHeap ---+      |  gcMode                  |
+------------+                    |      |  gcBgMarkStartWorkers    |
                                  |      +- STOP-THE-WORLD ---------+
+------------+                    |      |  finishsweep_m           |
| runtime.GC | --- TriggerCycle --+----> |  clearpools              |
+------------+                    |      +--------------------------+
                                  |      |  gcController.startCycle |
+------------+                    |      |  Phase: _GCmark          |
| sysmon     | --- TriggerTime ---+      |  gcMarkRootPrepare       |
+------------+                           |  gcBlackenEnabled = 1    |
|  forcegc.g |                           +- START-THE-WORLD --------+
+------------+                           |  Gosched/BackgroundMode  |
                                         +--------------------------+
```
