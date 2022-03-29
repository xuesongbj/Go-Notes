# 线程

![线程](./imgs/线程.png)

* `wakep`：唤醒MP执行任务。
* `startm`：新建或唤醒闲置M。
    * `newm`：新建M。
    * `mget`：获取闲置M。
* `newproc`：新建系统线程。
    * `mstart, mstart1`：线程入口函数。
* `schedule`：调度函数。
    * `findrunnable`：查找可用任务。
    * `stopm`：停止MP，进入闲置休眠状态。
