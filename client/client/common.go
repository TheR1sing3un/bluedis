package client

import (
	"os"
	"path/filepath"
)

var commandList = [][]string{
	{"SET", "key value", "STRING"},
	{"GET", "key", "STRING"},
	{"SETNX", "key seconds value", "STRING"},
	{"SETEX", "key value", "STRING"},
	{"GETSET", "key value", "STRING"},
	{"MSET", "[key value...]", "STRING"},
	{"MGET", "[key...]", "STRING"},
	{"APPEND", "key value", "STRING"},
	{"STREXISTS", "key", "STRING"},
	{"REMOVE", "key", "STRING"},
	{"EXPIRE", "key seconds", "STRING"},
	{"PERSIST", "key", "STRING"},
	{"TTL", "key", "STRING"},

	{"LPUSH", "key value [value...]", "LIST"},
	{"RPUSH", "key value [value...]", "LIST"},
	{"LPOP", "key", "LIST"},
	{"RPOP", "key", "LIST"},
	{"LINDEX", "key index", "LIST"},
	{"LREM", "key value count", "LIST"},
	{"LINSERT", "key BEFORE|AFTER pivot element", "LIST"},
	{"LSET", "key index value", "LIST"},
	{"LTRIM", "key start end", "LIST"},
	{"LRANGE", "key start end", "LIST"},
	{"LLEN", "key", "LIST"},
	{"LKEYEXISTS", "key", "LIST"},
	{"LVALEXISTS", "key value", "LIST"},
	{"LClear", "key", "LIST"},
	{"LExpire", "key seconds", "LIST"},
	{"LTTL", "key", "LIST"},

	{"HSET", "key field value", "HASH"},
	{"HSETNX", "key field value", "HASH"},
	{"HGET", "key field", "HASH"},
	{"HMSET", "[key field...]", "HASH"},
	{"HMGET", "[key...]", "HASH"},
	{"HGETALL", "key", "HASH"},
	{"HDEL", "key field [field...]", "HASH"},
	{"HKEYEXISTS", "key", "HASH"},
	{"HEXISTS", "key field", "HASH"},
	{"HLEN", "key", "HASH"},
	{"HKEYS", "key", "HASH"},
	{"HVALS", "key", "HASH"},
	{"HCLEAR", "key", "HASH"},
	{"HEXPIRE", "key seconds", "HASH"},
	{"HTTL", "key", "HASH"},

	{"SADD", "key members [members...]", "SET"},
	{"SPOP", "key count", "SET"},
	{"SISMEMBER", "key member", "SET"},
	{"SRANDMEMBER", "key count", "SET"},
	{"SREM", "key members [members...]", "SET"},
	{"SMOVE", "src dst member", "SET"},
	{"SCARD", "key", "key", "SET"},
	{"SMEMBERS", "key", "SET"},
	{"SUNION", "key [key...]", "SET"},
	{"SDIFF", "key [key...]", "SET"},
	{"SKEYEXISTS", "key", "SET"},
	{"SCLEAR", "key", "SET"},
	{"SEXPIRE", "key seconds", "SET"},
	{"STTL", "key", "SET"},

	{"ZADD", "key score member", "ZSET"},
	{"ZSCORE", "key member", "ZSET"},
	{"ZCARD", "key", "ZSET"},
	{"ZRANK", "key member", "ZSET"},
	{"ZREVRANK", "key member", "ZSET"},
	{"ZINCRBY", "key increment member", "ZSET"},
	{"ZRANGE", "key start stop", "ZSET"},
	{"ZREVRANGE", "key start stop", "ZSET"},
	{"ZREM", "key member", "ZSET"},
	{"ZGETBYRANK", "key rank", "ZSET"},
	{"ZREVGETBYRANK", "key rank", "ZSET"},
	{"ZSCORERANGE", "key min max", "ZSET"},
	{"ZREVSCORERANGE", "key max min", "ZSET"},
	{"ZKEYEXISTS", "key", "ZSET"},
	{"ZCLEAR", "key", "ZSET"},
	{"ZEXPIRE", "key", "ZSET"},
	{"ZTTL", "key", "ZSET"},

	{"MULTI", "Transaction start", "TRANSACTION"},
	{"EXEC", "Transaction end", "TRANSACTION"},
	{"PING"},
	{"AUTH"},
}

//终端命令历史记录文件路径
var history_fn = filepath.Join(os.TempDir(), ".liner_history")
