## goserver

goserver 旨在做一个传统的CS结构的服务器框架
目前主要用于游戏服务器开发
框架还在不断努力完善中，如果你对它感兴趣，请关注它的动态或者参与进来

## Features

* 组件通过package的概念统一管理(可以理解为win32下的dll)，由config来配置各个组件的特性参数
* goroutine通过Object进行包装以树型结构组织，Object间的通信通过command(内部是chan),主要是为了预防chan滥用、失控，从而造成各种死锁问题
* 提供了时间，任务，事务，计划工作，网络通讯，模块管理的内置组件
* 提供一套传统的游戏服务器架构(制作中...)

## 模块说明
* +core 核心模块
	* admin : http管理接口，主要提供一种外部可以操控进程的可能
	* basic : 基础的线程对象，封装对象间内部通讯；避免chan环锁现象，树形管理object
	* bulletin: 框架内建元素，提供通讯层的一些基础过滤器和通讯协议
	* cmdline: 自建命令行，给控制台进程提供一种命令模式
	* container: 框架用到的一些容器，队列，回收器，线程安全list，线程安全map
	* i18n: 国际化配置
	* logger: 日志接口
	* module: 业务模块管理，提供统一的心跳管理，模块通过注册挂载到管理器
	* mongo: mogodb相关配置
	* netlib: 通讯模块，支持TCP和WebSocket两种通讯方式
	* profile: 性能统计相关，用于辅助查找性能热点
	* schedule: 定时任务调度模块，用于周期job处理，如：每日凌晨4：00进行日志清理
	* signal: 信号管理模块,hook操作系统的信号进行回调处理，如：kill -2 PID
	* task: 线程模块，提供线程池、实名线程和独立线程多种模式
	* timer: 定时器，有别于go内置的timer；主要用于确保线程安全问题
	* transact: 分布式事务，基于二段提交实现，协调多节点配合完成一件原子性操作
	* utils: 工具接口
	* zk: zookeeper接口，用于分布式协调
* +srvlib core/netlib的扩展封装，提供常用的客户端session和服务端service管理，以及服务发现；进一步封装，使框架层达到拆箱即用
	* action 内置常用的包重定向和中转操作
	* handler 提供基本的session和service管理
	* protocol 内置协议定义
* +examples 示例程序
	* echoclient 回声客户端程序
	* echoserver 回声服务端程序
	* other timer和task使用示例
	* txserver1 分布式事务节点1
	* txserver2 分布式事务节点2
* +mmo 提供一套基本的服务器架构模板