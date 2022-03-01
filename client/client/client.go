package client

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/peterh/liner"
	"log"
	"os"
	"strings"
	"sync"
)

type Client struct {
	mu         sync.Mutex
	routerIp   string
	routerPort int
	conn       redis.Conn
	line       *liner.State
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
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", routerIp, routerPort))
	if err != nil {
		client.log("连接 %s:%d 失败: %v\n", routerIp, routerPort, err)
		return nil
	}
	client.conn = conn
	client.line = liner.NewLiner()
	//配置liner
	client.configureLiner()
	return client
}

func (c *Client) StartClient() {
	defer c.conn.Close()
	defer c.line.Close()
	defer c.writeLineHistory()
	prefix := fmt.Sprintf("%s:%d>", c.routerIp, c.routerPort)
	for {
		//接收消息
		cmd, err := c.line.Prompt(prefix)
		if err != nil {
			fmt.Println("输入错误:", err)
			return
		}
		//删掉多余空格
		cmd = strings.TrimSpace(cmd)
		if len(cmd) == 0 {
			continue
		}
		//记录到命令历史中
		c.line.AppendHistory(cmd)
		//将消息发过去
		if c.applyCommand(cmd) {
			//需要关闭客户端
			fmt.Println("bye")
			return
		}
	}
}

func (c *Client) applyCommand(cmd string) (exit bool) {
	exit = false
	if cmd == "quit" {
		return true
	}
	//解析出命令
	command, args := parseCommand(cmd)
	resp, err := c.conn.Do(command, args...)
	if err != nil {
		fmt.Printf("(error) %v\n", err)
		return
	}
	//响应命令
	switch reply := resp.(type) {
	case string:
		fmt.Println(reply)
	case []byte:
		fmt.Println(string(reply))
	case nil:
		fmt.Println("(nil)")
	case redis.Error:
		fmt.Printf("(error) %v\n", reply)
	case int64:
		fmt.Printf("(integer) %d\n", reply)
	default:
		return
	}
	return
}

func parseCommand(cmd string) (cmdType string, args []interface{}) {
	//以空格分割
	eles := strings.Split(cmd, " ")
	if len(cmd) == 0 {
		return "", nil
	}
	args = make([]interface{}, 0)
	for _, ele := range eles {
		if ele == "" {
			continue
		}
		//否则加入并且转为小写
		args = append(args, strings.ToLower(ele))
	}
	cmdType = fmt.Sprintf("%s", args[0])
	return cmdType, args[1:]
}

func (c *Client) configureLiner() {
	//设置Ctrl C退出确认
	c.line.SetCtrlCAborts(true)
	//设置命令自动补全(按Tab键)
	c.line.SetCompleter(func(line string) (res []string) {
		for _, cmd := range commandList {
			//当命令列表数组中每个数组的第一个字符串,也就是命令类型,和当前终端输入的如果匹配,就加入结果集
			if strings.HasPrefix(cmd[0], strings.ToUpper(line)) {
				res = append(res, strings.ToLower(cmd[0]))
			}
		}
		return
	})
	//初始化命令历史记录
	if file, err := os.Open(history_fn); err == nil {
		//先从文件中读取之前的历史记录
		c.line.ReadHistory(file)
		//关闭文件
		file.Close()
	}
}

//结束的时候进行命令历史的文件写入(下一次可以直接从文件中恢复)
func (c *Client) writeLineHistory() {
	if file, err := os.Create(history_fn); err == nil {
		//写到文件中
		c.line.WriteHistory(file)
		file.Close()
	} else {
		c.log("error writing history file: %v\n", err)
	}
}
