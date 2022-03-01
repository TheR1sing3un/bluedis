package main

import (
	"bluedis/router"
	"log"
)

var (
	address = []string{"localhost:8000", "localhost:8001", "localhost:8002"}
)

const (
	ip   = "0.0.0.0"
	port = 6380
)

//路由层
func main() {
	//创建router
	router, err := router.NewRouter(address, ip, port)
	if err != nil {
		log.Fatalf("router: 创建router错误,err: %v\n", err)
	}
	log.Printf("开始服务,当前地址是: %s:%d,kvserver组的地址是:\n%v\n", ip, port, address)
	//启动router
	router.StartRouter()
}
