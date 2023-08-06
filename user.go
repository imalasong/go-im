package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string

	conn net.Conn

	C chan string

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		conn:   conn,
		C:      make(chan string),
		server: server,
	}
	go user.Listener()
	return user
}

func (user *User) Online() {
	user.server.Lock.Lock()
	user.server.UserMap[user.Name] = user
	user.server.Lock.Unlock()
	user.server.BroadcastMsg(user, "用户上线")
}

func (user *User) OffOnline() {
	user.conn.Close()
	user.server.BroadcastMsg(user, "用户下线")
	user.server.Lock.Lock()
	delete(user.server.UserMap, user.Name)
	user.server.Lock.Unlock()

}
func (user *User) SendMsgToSelf(msg string) {
	user.conn.Write([]byte(msg))
}

func (user *User) DoMessage(msg string) {
	fmt.Println("收到消息:", msg)
	if len(msg) > 6 && msg[:7] == "rename|" {
		//修改名称
		newName := msg[7:]
		user.server.Lock.Lock()
		_, ok := user.server.UserMap[newName]
		if ok {
			user.SendMsgToSelf("名称:" + newName + ",已经存在")
			return
		} else {
			delete(user.server.UserMap, user.Name)
			user.Name = newName
			user.server.UserMap[newName] = user
			user.server.BroadcastMsg(user, "修改名称成功，newName="+newName)
		}
		user.server.Lock.Unlock()
	} else {
		//广播
		user.server.BroadcastMsg(user, msg)
	}

}

func (user *User) Listener() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
