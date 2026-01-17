package main

import (
	"fmt" // 格式化输出
	"io"
	"net"  // 网络操作
	"sync" // 同步包
)

// Server 定义服务器结构
type Server struct {
	Ip        string // 服务器IP地址
	Port      int    // 服务器端口号
	OnlineMap map[string]*User
	mapLock   sync.RWMutex // 保护OnlineMap的互斥锁，并发安全
	Message   chan string
}

// NewServer 创建新的服务器实例
func NewServer(ip string, port int) *Server { // * 有两层含义, 1. 表示返回一个指向 Server 结构体的指针, 2. 表示 Server 结构体的实例, 当前表示定义一个指针
	return &Server{ // & 表示取地址
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User), // make() Go的内置函数, 专门用于创建引用类型
		Message:   make(chan string),
	}
}

// Start 启动服务器
func (this *Server) Start() {
	// 1. 创建 TCP 监听 : (创建一个 TCP 协议的监听器 (Listener)，并让这个监听器在指定的IP:端口上执行「监听」动作)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil { // 如果监听失败, 打印错误信息并返回
		fmt.Println("Listen err:", err)
		return
	}
	defer listener.Close() // 关闭监听 : 确保在函数返回前关闭监听, 释放资源

	// 2️⃣ 启动监听消息 Message channel 的goroutine, 相当于server在消费消息
	go this.ListenMessage()

	fmt.Println("Server start at:", fmt.Sprintf("%s:%d", this.Ip, this.Port))

	// 2. 循环接受客户端连接 :
	for {
		conn, err := listener.Accept() // 接受客户端连接 : (阻塞式等待, 直到有客户端连接到服务器)
		if err != nil {                // 如果接受连接失败, 打印错误信息并继续循环
			fmt.Println("listener accept err:", err)
			continue
		}

		// 3. 3️⃣启动 goroutine 处理客户端连接, 并将信息存到 message 这个 channel 中, 相当于创造消息
		go this.Handler(conn)
	}
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("链接建立成功,客户端地址 : ", conn.RemoteAddr()) // 打印新连接的远程地址
	// 创建用户对象
	user := NewUser(conn, this)

	// 被Online方法替换掉,用户上线, 加入在线用户列表:
	user.OnLine()

	// 启动一个 goroutine, 专门处理客户端发送的消息, 然后将客户端发送的消息投送到 Message 中, 进行广播
	go func() {
		buf := make([]byte, 4096) // 创建一个 4KB 的字节缓冲区, 用于存储客户端发送的消息, 使用切片实现这个功能
		for {
			n, err := conn.Read(buf)
			if n == 0 { // 客户端主动下线 → n=0
				user.OffLine() // 被OffLine方法替换掉,用户下线, 从在线列表中删除
				conn.Close()   // 关闭连接
				return
			}
			if err != nil && err != io.EOF { // 其他错误, 打印错误信息并返回
				fmt.Println("conn.Read err:", err)
				return
			}
			// 一切正常, 接收并处理消息
			msg := string(buf[:n-1]) // 去掉换行符
			this.Message <- msg      // 将消息投递到全局通道 Message 中
		}
	}()

	// 阻塞当前 handler goroutine, 防止退出:
	select {}
}

// 作用是投递消息到全局通道 Message
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s:%s", user.Addr, user.Name, msg)
	this.Message <- sendMsg
}

// ListenMessage 监听全局Message消息 channel, 将消息发送给所有在线用户:
// 作用是用来监听Message中的消息, 如果有新消息, 就传递给不同 user 的 channel 中, 从而实现广播
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		// 遍历所有在线用户, 发送消息
		this.mapLock.RLock()
		for _, cli := range this.OnlineMap { // 专门用来遍历 map 的标准写法
			cli.C <- msg // 将消息传递给所有 Client 的 channel 中
		}
		this.mapLock.RUnlock()
	}
}
