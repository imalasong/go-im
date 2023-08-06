package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
)

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "need conneting server‘s ip")
	flag.IntVar(&serverPort, "port", 8888, "need conneting server‘s port")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	client.Run()
}

type Client struct {
	ip string

	port int

	runing bool
}

func NewClient(ip string, port int) *Client {

	client := &Client{
		ip:   ip,
		port: port,
	}
	return client
}

func (client *Client) Run() {

	address := fmt.Sprintf("%s:%d", client.ip, client.port)

	conn, err := net.Dial("tcp", address)

	if err != nil {
		fmt.Printf("连接[%s]失败\n", err)
		return
	}
	fmt.Printf("连接[%s]成功ln\n", address)

	client.runing = true

	go client.Menu(conn)

	client.ReceiveMsgHandler(conn)
}

func (client *Client) Menu(conn net.Conn) {
	for {
		opera := client.UserOperation()
		switch opera {
		case 1:
			client.SendMsgProccesss(conn)
		case 2:
			client.ChangeNameProccesss(conn)
		}
	}
}

func (client *Client) SendMsgProccesss(conn net.Conn) {
	fmt.Println("请输入你要发送的内容：")

	var msg string

	_, err := fmt.Scanln(&msg)
	if err != nil {
		fmt.Println("输入信息异常,", err)
		return
	}

	_, err = conn.Write([]byte(msg))

	if err != nil {
		fmt.Println("发送消息异常,", err)
	}

}

func (client *Client) ChangeNameProccesss(conn net.Conn) {
	fmt.Println("请输入你的新名字：")

	var msg string

	_, err := fmt.Scanln(&msg)
	if err != nil {
		fmt.Println("输入信息异常,", err)
		return
	}

	conn.Write([]byte("rename|" + msg))
}

func (client *Client) UserOperation() int {
	for {
		fmt.Println("1、发送消息\n2、修改名称")
		var opera int
		_, err := fmt.Scanln(&opera)
		if err != nil {
			fmt.Println("输入格式错误❌")
			continue
		}
		if opera != 1 && opera != 2 {
			fmt.Println("输入错误❌")
			continue
		}
		return opera
	}

}

func (client *Client) ReceiveMsgHandler(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)

		len, err := conn.Read(buffer)
		if len == 0 {
			fmt.Println("服务端主动断开连接")
			runtime.Goexit()
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("读取消息异常")
			continue
		}
		msg := strings.TrimSpace(string(buffer[0:len]))
		fmt.Println("收到消息:", msg)
	}
}
