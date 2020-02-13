package config

import (
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"os"
	"strings"
)

var (
	Date    string
	Version string
	Commit  string
)

type RedisConfig struct {
	Enable bool   `json:"enable"`
	Addr   string `json:"addr"`
	Auth   string `json:"auth"`
	Db     int    `json:"database"`
}

func RightAddr(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func GetAlias() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return libol.GenToken(13)
}

func init() {
	libol.Info("Config: Version %s", Version)
	libol.Info("Config: Date %s", Date)
	libol.Info("Config: Commit %s", Commit)
}