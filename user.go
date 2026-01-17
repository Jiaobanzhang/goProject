package main

import (
	"net"
)

// User 定义用户结构
type User struct {
	Name string      // 用户名称
	Addr string      // 用户地址
	C    chan string // 用户消息通道
	conn net.Conn    // 用户连接
}

// 构造器 : NewUser 创建新的用户实例
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// 3️⃣ 启动监听用户 channel 的 goroutine, 相当于 Client 在消费消息
	go user.ListenMessage()
	return user
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n")) // net.Conn 只能传输 二进制字节流[]byte
	}
}
