package main

import (
	"bluedis/kvraft"
	"bluedis/raft"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

var address = map[int]string{
	0: "localhost:8000",
	1: "localhost:8001",
	2: "localhost:8002",
}

var serverIp string
var serverPort int
var serverNum int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server的ip")
	flag.IntVar(&serverPort, "port", 8000, "server的端口")
	flag.IntVar(&serverNum, "id", 0, "该server在Raft组中的编号")
}

//启动一个服务端(接收client请求)
func main() {
	//解析参数
	flag.Parse()
	persister := raft.MakePersister()
	kvServer := kvraft.NewKVServer(serverNum, -1)
	//在rpc中进行注册
	err := rpc.RegisterName("KVServer", kvServer)
	if err != nil {
		return
	}
	//监听自己的ip:port
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", address[serverNum])
	if err != nil {
		log.Fatalf("监听自己的server: %s失败,error: %v\n", address[serverNum], err)
		return
	}
	fmt.Printf("监听成功: %v\n", listener.Addr())
	defer listener.Close()
	serverEnds := make([]*rpc.Client, len(address))
	serverEnds[serverNum] = nil
	delete(address, serverNum)
	//启动http服务(为了rpc调用)
	go http.Serve(listener, nil)
	//等待其他节点上线
	fmt.Println("等待其他节点上线")
	for len(address) > 0 {
		for index, server := range address {
			//和server建立rpc连接
			client, err := rpc.DialHTTP("tcp", server)
			if err != nil {
				fmt.Printf("连接到server: %v失败,error: %v\n", server, err)
			} else {
				//连接成功则加入到serverEnds中
				fmt.Printf("节点:[%v]连接成功\n", server)
				serverEnds[index] = client
				delete(address, index)
			}
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("已连接所有节点,开始服务\n")
	//等节点都上线了,就开启KVServer
	kvServer.StartServer(persister, serverEnds)
}
