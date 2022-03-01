package main

import (
	"bluedis/client/client"
	"flag"
	"fmt"
	"github.com/peterh/liner"
	"io/ioutil"
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
	line := liner.NewLiner()
	defer line.Close()
	//创建client
	client := client.NewClient(ip, port)
	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}
	banner, err := ioutil.ReadFile("./client/banner.txt")
	if err != nil {
		fmt.Printf("open banner file error: %v\n", err)
	} else {
		fmt.Println(string(banner))
	}
	client.StartClient()
}
