package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	UserMap map[string]*User

	Lock sync.RWMutex

	MessageChannel chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{Ip: ip, Port: port, UserMap: make(map[string]*User), MessageChannel: make(chan string)}
	return server
}

func (this *Server) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("server start fail", err)
		return
	}

	fmt.Println("server start success！")

	go this.ListenerMessage()

	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("connector exception,", err)
			continue
		}
		go this.Handler(conn)
	}
}

func (this *Server) BroadcastMsg(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + msg

	this.MessageChannel <- sendMsg
}

func (this *Server) ListenerMessage() {
	for {
		msg := <-this.MessageChannel
		this.Lock.Lock()
		for _, user := range this.UserMap {
			user.C <- msg
		}
		this.Lock.Unlock()
	}
}

func (this *Server) Handler(conn net.Conn) {

	user := NewUser(conn, this)
	user.Online()

	heartbear := make(chan bool)

	go func() {
		buff := make([]byte, 1024)
		for {
			len, err := conn.Read(buff)
			if len == 0 {

				user.OffOnline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("read data error,", err)
				return
			}
			heartbear <- true
			receivemsg := strings.TrimSpace(string(buff[0:len]))
			user.DoMessage(receivemsg)
		}
	}()

	//超时剔除
	for {
		select {
		case <-heartbear:
			//
		case <-time.After(time.Second * 10):
			user.SendMsgToSelf("您被踢了\n")
			close(user.C)
			user.OffOnline()
			return
		}
	}
}

func (this *Server) Handler1(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()
	fmt.Println("client connecting success:", remoteAddr)
	conn.Write([]byte("hello\n"))

	for {
		var readB []byte = make([]byte, 1024)
		len, err := conn.Read(readB)
		if len == 0 {
			conn.Close()
			return
		}
		if err != nil {
			fmt.Println("read data error,", err)
			continue
		}
		receivemsg := strings.TrimSpace(string(readB[0 : len-1]))
		fmt.Println("receive data,len=", len, ",data=", receivemsg)
		if receivemsg == "bye" {
			conn.Write([]byte("goodbye\n"))
			closeErr := conn.Close()
			if closeErr != nil {
				fmt.Printf("connector %s close error,%v\n", remoteAddr, closeErr)
			} else {
				fmt.Printf("connector %s close success\n", remoteAddr)
				return
			}

		} else {
			conn.Write(readB)
		}
	}

}
