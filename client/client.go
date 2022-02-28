package main

import (
	"bluedis/client/client"
	"flag"
	"fmt"
)

var (
	ip   string
	port int
)

func init() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "router的ip地址")
	flag.IntVar(&port, "port", 6380, "router的端口")
}

func main() {
	flag.Parse()
	//创建client
	client := client.NewClient(ip, port)
	if client == nil {
		fmt.Println(">>>>>>>>>>连接服务器失败>>>>>>>>>>")
		return
	}
	fmt.Println(">>>>>>>>>>连接服务器成功>>>>>>>>>>")
	client.StartClient()
}
