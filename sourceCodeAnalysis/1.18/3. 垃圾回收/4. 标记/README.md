# 标记

```go

+--------------------------+
| gcStart                  |
+--------------------------+
|  === stop-the-world ===  |    +--- dedicatedMarkWorkersNeeded = ?
+--------------------------+    |
|  startCycle             -|----+
+--------------------------+    |
|  Phase: _GCmark          |    +--- fractionalUtilizationGoal = ?
+--------------------------+
|  gcMarkRootPrepare       |
+--------------------------+
|  gcBlackenEnabled = 1    |
+--------------------------+            
|  === start-the-world === |
+--------------------------+
           |  
         allp
           |
           v
+----------+----------------------+------------------------+
| schedule | findRunnableGCWorker | findrunnable (nothing) |
+----------+----------------------+------------------------+
                      |                        |
            dedicated, fractional            idle
                      |                        |
                      +------------+-----------+
                                   |
                                   v               +--- p.gcw
                       +----------------------+    |
                       | gcMarkWorkAvailable -|----+--- work.full
                       +----------------------+    |
                                   |               +--- work.markrootJobs
                                   v
                          +----------------+
                          | gcBgMarkWorker |  (worker groutine)
                          +----------------+
+----------------+        |  gcDrain       |
| mutator        |        +----------------+
+----------------+        |  gcMarkDone    |
|  mallocgc      |        +----------------+
+----------------+                 |
|  gcAssistAlloc |                 v            +------------+------------+
+----------------+           +----------+       | markroot   | scanblock  |
|  gcDrain      -|---------> | gcDrain -|-----> +------------+------------+ 
+----------------+           +----------+       | scanobject | greyobject |
|  gcMarkDone    |                 |            +------------+------------+
+----------------+                 v
                      +-------------------------+
                      | gcMarkDone              |
                      +-------------------------+
                      |  === stop-the-world === |
                      +-------------------------+
                      |  gcBlackenEnabled = 0   |
                      +-------------------------+
                      |  gcController.endCycle -|---- nextTriggerRatio
                      +-------------------------+
                      |  gcMarkTermination      |
                      +-------------------------+
                                   |
                                   v
                    +----------------------------+
                    | gcMarkTermination          |
                    +----------------------------+
                    |  Phase: _GCmarktermination |
                    +----------------------------+     +--------+
                    |  gcMark                   -|---> | gcMark |
                    +----------------------------+     +--------+
                    |  Phase: _GCoff             |
                    +----------------------------+
                    |  gcSweep                   |
                    +----------------------------+
                    |  === start-the-world ===   |
                    +----------------------------+
                         
```