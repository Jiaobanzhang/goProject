package main

import (
	"net"
)

// User 定义用户结构
type User struct {
	Name   string      // 用户名称
	Addr   string      // 用户地址
	C      chan string // 用户消息通道
	conn   net.Conn    // 用户连接
	server *Server     // 用户所属服务器
}

// 构造器 : NewUser 创建新的用户实例
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server, // 关联到当前服务器
	}

	// 3️⃣ 启动监听用户 channel 的 goroutine, 相当于 Client 在消费消息
	go user.ListenMessage()
	return user
}

// 作用是监听用户 channel, 并将消息写入到用户连接中
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n")) // net.Conn 只能传输 二进制字节流[]byte
	}
}

// Online 用户上线业务 :
func (this *User) OnLine() {
	// 1. 用户上线, 加入到在线列表中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 2. 广播用户上线消息:
	this.server.BroadCast(this, "已上线") // 只要有新用户上线了, 就广播一次
}

// OffLine 用户下线业务 :
func (this *User) OffLine() {
	// 1. 用户下线, 从在线列表中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 2. 广播用户下线消息:
	this.server.BroadCast(this, "已下线") // 只要有用户下线了, 就广播一次
}

// DoMessage 处理用户消息业务 :
func (this *User) DoMessage(msg string) {
	// 现在处理 "who" 指令 :
	if msg == "who" {
		this.server.mapLock.RLock() // 加读锁, 防止在遍历过程中, 有用户上线或下线
		this.SendMsg("当前在线用户列表:")
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.RUnlock() // 遍历完成后, 解锁
	} else {
		this.server.BroadCast(this, msg) // 广播用户消息
	}
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n")) // net.Conn 只能传输 二进制字节流[]byte
}
