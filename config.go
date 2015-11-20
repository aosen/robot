// Package etc_config implements config initialization of one spider.
package robot

import (
	"github.com/aosen/goconfig"
	"github.com/aosen/utils"
	"os"
)

// Config is a config singleton object for one spider.
var conf *goconfig.Config
var path string

// Configpath gets default config path like "WD/etc/main.conf".
func configpath() string {
	//wd, _ := os.Getwd()
	wd := os.Getenv("GOPATH")
	if wd == "" {
		panic("GOPATH is not setted in env.")
	}
	logpath := wd + "/etc/"
	filename := "robot.conf"
	err := os.MkdirAll(logpath, 0755)
	if err != nil {
		panic("logpath error : " + logpath + "\n")
	}
	return logpath + filename
}

// StartConf is used in Spider for initialization at first time.
func StartConf(configFilePath string) *goconfig.Config {
	if configFilePath != "" && !utils.IsFileExists(configFilePath) {
		panic("config path is not valiad:" + configFilePath)
	}

	path = configFilePath
	return Conf()
}

// Conf gets singleton instance of Config.
func Conf() *goconfig.Config {
	if conf == nil {
		if path == "" {
			path = configpath()
		}
		conf = goconfig.NewConfig().Load(path)
	}
	return conf
}
