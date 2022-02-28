package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Client struct {
	mu         sync.Mutex
	routerIp   string
	routerPort int
	conn       net.Conn
}

const debug = true

func (c *Client) log(format string, v ...interface{}) {
	if debug {
		log.Printf("client: %v\n", fmt.Sprintf(format, v...))
	}
}

func NewClient(routerIp string, routerPort int) *Client {
	client := new(Client)
	client.routerIp = routerIp
	client.routerPort = routerPort
	//开始连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", routerIp, routerPort))
	if err != nil {
		client.log("连接 %s:%d 失败: %v\n", routerIp, routerPort, err)
		return nil
	}
	client.conn = conn
	return client
}

func (c *Client) StartClient() {
	defer c.conn.Close()
	go c.dealResp()
	for {
		//接收消息
		var cmd string
		fmt.Print("->")
		reader := bufio.NewReader(os.Stdin)
		cmd, _ = reader.ReadString('\n')
		if cmd == "exit" {
			fmt.Println("已退出,欢迎下次使用")
			return
		}
		//将消息发送过去
		if len(cmd) != 0 {
			sendCmd := cmd
			if !c.sendCommand(sendCmd) {
				fmt.Println("消息发送失败")
				return
			}
		}
	}
}

func (c *Client) sendCommand(cmd string) bool {
	_, err := c.conn.Write([]byte(cmd))
	if err != nil {
		c.log("conn.Write [%v] error: %v", cmd, err)
		return false
	}
	//if err != nil {
	//	c.log("conn.Read error: %v", err)
	//	return "", false
	//}
	return true
}

func (c *Client) dealResp() {
	io.Copy(os.Stdout, c.conn)
}
