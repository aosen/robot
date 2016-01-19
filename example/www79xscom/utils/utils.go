/*
Author: Aosen
Date: 2016-01-19
QQ: 316052486
Desc: 工具箱
*/
package utils

import (
	"github.com/aosen/robot"
	"github.com/aosen/utils"
)

const (
	BaseUrl string = "http://www.79xs.com"
	GirlUrl string = "/book/LC/165.aspx"
	GIRL    string = "女生"
	BOY     string = "男生"
)

//停止标志
var Stop chan bool = make(chan bool)

func LoadConf(path string) (settings map[string]string) {
	//生成配置文件对象,加载配置文件
	config := utils.NewConfig().Load(path)
	return config.GlobalContent()
}

func InitRequest(url string, meta map[string]string, cb func(*robot.Page)) *robot.Request {
	return &robot.Request{
		Url:      url,
		RespType: "html",
		Meta:     meta,
		CallBack: cb,
	}
}

func Stopspider() {
	close(Stop)
}
