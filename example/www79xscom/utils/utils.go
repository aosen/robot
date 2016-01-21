/*
Author: Aosen
Date: 2016-01-19
QQ: 316052486
Desc: 工具箱
*/
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"

	"github.com/aosen/goutils"
	"github.com/aosen/robot"
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
	config := goutils.NewConfig().Load(path)
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

//map[string]string 拷贝
func MapCopy(src map[string]string) map[string]string {
	dest := make(map[string]string)
	for k, v := range src {
		dest[k] = v
	}
	return dest
}

//下载图片
func DownloadImage(url, path string) (filename string, err error) {
	var res *http.Response
	var file *os.File
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(url))
	cipherStr := md5Ctx.Sum(nil)
	name := hex.EncodeToString(cipherStr)
	pathname := path + name + ".jpg"
	res, err = http.Get(url)
	defer res.Body.Close()
	os.MkdirAll(path, os.ModePerm)
	file, err = os.Create(pathname)
	defer file.Close()
	io.Copy(file, res.Body)
	filename = name + ".jpg"
	return
}
