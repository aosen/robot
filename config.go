// Package etc_config implements config initialization of one spider.
package robot

import (
	"os"

	"github.com/aosen/goutils"
)

// Config is a config singleton object for one spider.
var conf *goutils.Config
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
func StartConf(configFilePath string) *goutils.Config {
	if configFilePath != "" && !goutils.IsFileExists(configFilePath) {
		panic("config path is not valiad:" + configFilePath)
	}

	path = configFilePath
	return Conf()
}

// Conf gets singleton instance of Config.
func Conf() *goutils.Config {
	if conf == nil {
		if path == "" {
			path = configpath()
		}
		conf = goutils.NewConfig().Load(path)
	}
	return conf
}
