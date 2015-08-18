## goserver

goserver 旨在做一个传统的CS结构的服务器框架，支持分布式部署
目前主要用于游戏服务器开发
框架在不断努力完善中

## Features

* 组件通过package的概念统一管理(可以理解为win32下的dll)，由config来配置各个组件的特性参数
* goroutine通过Object进行包装以树型结构组织，Object间的通信通过command(内部是chan),支持chan自动增长;主要是为了预防chan泛滥、无约束，从而造成一些莫名或者不可控的死锁问题
* 提供了定时器(timer)，任务(task)，分布式事务(transact)，计划工作(schedule)，网络通讯(netlib)，模块管理(module)等基础的内置组件
* 提供session管理、service注册，查找等分布式相关的基础库(srvlib)
* 提供一套传统的游戏服务器架构(mmo)
