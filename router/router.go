package router

import (
	"bluedis/kvraft"
	"fmt"
	"github.com/tidwall/redcon"
	"log"
	"net/rpc"
	"strings"
	"sync"
)

type CommandType string

const (
	GET    = "get"
	APPEND = "APPEND"
	SET    = "set"
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
	return router, nil
}

func (r *Router) StartRouter() {
	//监听本机的端口
	address := fmt.Sprintf("%s:%d", r.ip, r.port)
	err := redcon.ListenAndServe(address,
		func(conn redcon.Conn, cmd redcon.Command) {
			r.log("接收到address: %v,command: %v", cmd, conn.Context())
			r.handleCmd(conn, cmd)
		},
		func(conn redcon.Conn) bool {
			return r.acceptConn(conn)
		},
		func(conn redcon.Conn, err error) {
			r.closeConn(conn, err)
		})
	if err != nil {
		r.log("listener出错,error: %v", err)
	}
}

func (r *Router) handleCmd(conn redcon.Conn, cmd redcon.Command) {
	//将命令类型转为小写,并判断命令类型
	switch strings.ToLower(string(cmd.Args[0])) {
	case SET:
		//若是set命令
		//1.判断参数个数是否正确
		if len(cmd.Args) != 3 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		//2.若正确,则应用命令
		r.clerk.Put(string(cmd.Args[1]), string(cmd.Args[2]))
		conn.WriteString("OK")
	case GET:
		//1.判断命令是否合规
		if len(cmd.Args) != 2 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		//2.若正确,则应用命令
		value := r.clerk.Get(string(cmd.Args[1]))
		if value == "" {
			conn.WriteNull()
		} else {
			conn.WriteBulk([]byte(value))
		}
	default:
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
	}
}

func (r *Router) acceptConn(conn redcon.Conn) bool {
	return true
}

func (r *Router) closeConn(conn redcon.Conn, err error) {
}
