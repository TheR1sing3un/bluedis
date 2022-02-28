package router

import (
	"bluedis/kvraft"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"strings"
	"sync"
)

type CommandType string

const (
	GET    = "GET"
	APPEND = "APPEND"
	SET    = "SET"
)

type Router struct {
	ip    string //ip地址
	port  int    //端口号
	clerk *kvraft.Clerk
	mu    sync.Mutex
}

const debug = true

func (r *Router) log(format string, v ...interface{}) {
	if debug {
		log.Printf("router: %v\n", fmt.Sprintf(format, v...))
	}
}

func NewRouter(serversAddress []string, ip string, port int) (*Router, error) {
	router := &Router{}
	router.ip = ip
	router.port = port
	serverEnds := make([]*rpc.Client, len(serversAddress))
	i := 0
	for _, address := range serversAddress {
		client, err := rpc.DialHTTP("tcp", address)
		if err != nil {
			sprintf := fmt.Sprintf("连接server: %s失败,error: %v\n", address, err)
			return nil, fmt.Errorf(sprintf)
		}
		serverEnds[i] = client
		i++
	}
	clerk := kvraft.MakeClerk(serverEnds)
	router.clerk = clerk
	fmt.Println("clientEnds:", serverEnds)
	return router, nil
}

func (r *Router) StartRouter() {
	//监听本机的端口
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", r.ip, r.port))
	if err != nil {
		log.Fatalf("监听本机: %s:%d失败,error: %v", r.ip, r.port, err)
	}
	defer listener.Close()
	//开始处理消息
	for {
		//接收连接
		conn, err := listener.Accept()
		//若接收错误,跳过该连接,继续
		if err != nil {
			r.log("连接错误,error: %v", err)
			continue
		}
		//处理连接
		go r.handleConn(conn)
	}
}

func (r *Router) handleConn(conn net.Conn) {
	go func() {
		//建立缓存区,用于接收命令
		buf := make([]byte, 4096)
		//不断接收命令
		for {
			//从conn中读取客户端发的消息
			n, err := conn.Read(buf)
			//当n == 0的时候,代表已经不再有数据接收了
			if n == 0 {
				return
			}
			if err != nil && err != io.EOF {
				r.log("[%v]: conn read err: %v", conn.RemoteAddr().String(), err)
				return
			}
			fmt.Println("buf:", buf)
			//提取消息(取0到n-2个,因为最后一个是'\n')
			cmd := string(buf[:n-2])
			fmt.Println("cmd:", cmd)
			//处理命令
			go r.applyCommand(cmd, conn)
		}
	}()

}

func (r *Router) applyCommand(cmd string, conn net.Conn) {
	cmds := strings.Split(cmd, " ")
	if len(cmds) < 1 {
		r.log("[%v]: invalid command: %v", conn.RemoteAddr().String(), cmd)
		return
	}
	fmt.Printf("type:%v;key:%v;\n", cmds[0], cmds[1])
	fmt.Println("keyLen:", len(cmds[1]))
	operation := cmds[0]
	operation = strings.ToUpper(operation)
	var reply string
	switch operation {
	case GET:
		key := cmds[1]
		value := r.clerk.Get(key)
		reply = value
	case SET:
		key := cmds[1]
		value := cmds[2]
		fmt.Println("value:", value)
		r.clerk.Put(key, value)
		reply = "ok"
	case APPEND:
		key := cmds[1]
		value := cmds[2]
		fmt.Println("value:", value)
		r.clerk.Append(key, value)
		reply = "ok"
	}
	//将reply写回到客户端
	_, err := conn.Write([]byte(reply + "\n"))
	if err != nil {
		r.log("[%v]: 向该conn写数据失败,data: %v,error: %v", conn.RemoteAddr().String(), reply, err)
	}
}
