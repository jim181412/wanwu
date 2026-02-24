package config

import (
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	_c *Config
)

type Config struct {
	Server ServerConfig `json:"server" mapstructure:"server"`
	Log    LogConfig    `json:"log" mapstructure:"log"`
	DB     db.Config    `json:"db" mapstructure:"db"`
	Redis  redis.Config `json:"redis" mapstructure:"redis"`
}

type ServerConfig struct {
	GrpcEndpoint   string `json:"grpc_endpoint" mapstructure:"grpc_endpoint"`
	MaxRecvMsgSize int    `json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

type DBConfig struct {
	Name string `json:"name" mapstructure:"name"`
}

const (
	ScopeTypeStr_PRIVATE  = "private"
	ScopeTypeStr_PUBLIC   = "public"
	ModelScopeTypePrivate = "1"
	ModelScopeTypePublic  = "2"
	ModelScopeTypeOrg     = "3"
)

func LoadConfig(in string) error {
	_c = &Config{}
	return util.LoadConfig(in, _c)
}

func Cfg() *Config {
	if _c == nil {
		log.Panicf("cfg nil")
	}
	return _c
}
