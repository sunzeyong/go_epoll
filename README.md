# 使用epoll编写网络服务
talk is cheap, show me code!  
让八股文变成代码，实践epoll使用方式。  
server端全程使用系统调用，不使用go的任何封装  

## 如何运行
```
// server端
// 没有使用go run main.go是因为main包默认不加载其他包，会出现 undefined: xxx报错。
go run .

// client端
go run .
```

## 设计思路
### 没有epoll时的服务端设计
1. 创建监听socket
2. Bind ip和端口
3. listen获取链接socket
4. 处理socket中数据 读取或者写入

### 使用epoll后的调整
1. 在上述第三步前，创建epoll （epoll_create）
2. 在第三步中，将监听得到的socket保存在epoll中 (epoll_ctl)
3. 第四步修改为，起goroutine获取epoll中的事件，处理事件 (epoll_wait)

# 基于epoll实现Reactor
reactor模式也叫dispatcher模式，在收到IO多路复用监听事件后，根据事件类型分配给某个进程/线程处理。如果每次分配都创建新的线程，创建销毁过程会消耗cpu，所以这里一般使用线程池来处理。

# todo
- [ ] select和poll实现server端
- [ ] 添加压测数据对比

