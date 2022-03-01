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
	GET  = "get"
	PING = "ping"
	SET  = "set"
	QUIT = "quit"
	AUTH = "auth"
)

type Router struct {
	ip    string //ip地址
	port  int    //端口号
	clerk *kvraft.Clerk
	mu    sync.Mutex
	conns map[string]*CliConn
}

type CliConn struct {
	conn  redcon.Conn
	state bool //是否可通过(当验证了密码的时候,就改为可通过)
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
	router.conns = make(map[string]*CliConn)
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
			var cmdStr string
			for _, arg := range cmd.Args {
				cmdStr += string(arg) + " "
			}
			r.log("接收到address: %v,command: %v", conn.RemoteAddr(), cmdStr)
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
	case AUTH:
		//1.验证格式
		if len(cmd.Args) != 2 && len(cmd.Args) != 3 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		//2.验证密码
		if (len(cmd.Args) == 2 && string(cmd.Args[1]) == password) || (len(cmd.Args) == 3 && string(cmd.Args[2]) == password) {
			//验证成功
			if cli, ok := r.conns[conn.RemoteAddr()]; ok {
				cli.state = true
			} else {
				r.conns[conn.RemoteAddr()] = &CliConn{conn, true}
			}
			conn.WriteString("OK")
			return
		}
		//3.密码错误
		conn.WriteString("(error) WRONGPASS invalid username-password pair")

	default:
		if cli, ok := r.conns[conn.RemoteAddr()]; ok && cli.state == false {
			//即还未验证密码
			conn.WriteString("(error) NOAUTH Authentication required.")
			return
		}
		switch strings.ToLower(string(cmd.Args[0])) {
		case QUIT:
			conn.WriteString("OK")
			conn.Close()
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
		case PING:
			//1.判断命令是否合规
			if len(cmd.Args) != 1 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
				return
			}
			//2.返回PONG
			conn.WriteString("PONG")
		default:
			conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
		}
	}

}

//当连接接收的时候调用
func (r *Router) acceptConn(conn redcon.Conn) bool {
	r.log("接收到连接: %v\n", conn.RemoteAddr())
	r.mu.Lock()
	defer r.mu.Unlock()
	//加入到连接列表中
	r.conns[conn.RemoteAddr()] = &CliConn{conn, !pwdAble}
	return true
}

//当连接被关闭的时候被调用
func (r *Router) closeConn(conn redcon.Conn, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.conns[conn.RemoteAddr()]; ok {
		delete(r.conns, conn.RemoteAddr())
	}
	r.log("关闭连接: %v\n", conn.RemoteAddr())
}
