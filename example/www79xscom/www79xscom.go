/*
Author: Aosen
Data: 2016-01-15
QQ: 316052486
Desc:
爬取www.79xs.com的所有小说内容，一键式智能爬取及更新
将spider.conf配置文件放在可执行文件的同层
爬虫先从主页www.79xs.com爬取, 提取分类目录并设置回调
抓取分类目录中的小说目录，提取二级分类，小说标题，并设置小说简介回调
抓取小说简介并设置章节回调
抓取章节并设置小说内容回调
抓取小说内容
如果抓到与数据库中大量相同的url 则退出, 阀值 100
*/

package main

import (
	"log"

	"github.com/aosen/robot"
	"github.com/aosen/robot/downloader"
	"github.com/aosen/robot/example/www79xscom/pipeline"
	"github.com/aosen/robot/example/www79xscom/process"
	"github.com/aosen/robot/example/www79xscom/utils"
	"github.com/aosen/robot/resource"
	"github.com/aosen/robot/scheduler"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	spidername := "79xs"
	start_url := utils.BaseUrl
	//加载配置文件
	settings := utils.LoadConf("conf/spider.conf")
	//获取数据库连接信息
	dbinfo, ok := settings["DBINFO"]
	if !ok {
		log.Fatalf("please insert dbinfo in spider.conf")
	}

	//爬虫初始化
	options := robot.SpiderOptions{
		TaskName:      spidername,
		PageProcesser: process.NewWww79xsComProcessor(),
		Downloader:    downloader.NewHttpDownloader("text/html; charset=gb2312"),
		Scheduler:     scheduler.NewMysqlScheduler(spidername, dbinfo),
		//Scheduler: scheduler.NewQueueScheduler(false),
		Pipelines: []robot.Pipeline{pipeline.NewPipelineMySQL(dbinfo)},
		//设置资源管理器，资源池容量为100
		ResourceManage: resource.NewSpidersPool(100, nil),
	}

	sp := robot.NewSpider(options)
	//增加根url
	sp.AddRequest(utils.InitRequest(start_url, map[string]string{
		"handler": "mainParse",
	}))
	go sp.Run()
	<-utils.Stop
	sp.Close()
}
