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
	"database/sql"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/mlog"
	"github.com/aosen/robot"
	"github.com/aosen/robot/downloader"
	"github.com/aosen/robot/resource"
	"github.com/aosen/robot/scheduler"
	"github.com/aosen/utils"
	_ "github.com/go-sql-driver/mysql"
)

const (
	baseurl  string = "http://www.79xs.com"
	girl_url string = "/book/LC/165.aspx"
	GIRL     string = "女生"
	BOY      string = "男生"
)

//停止标志
var stop chan bool = make(chan bool)

func loadconf(path string) (settings map[string]string) {
	//生成配置文件对象,加载配置文件
	config := utils.NewConfig().Load(path)
	return config.GlobalContent()
}

func initrequest(url string, meta map[string]string, cb func(*robot.Page)) *robot.Request {
	return &robot.Request{
		Url:      url,
		RespType: "html",
		Meta:     meta,
		CallBack: cb,
	}
}

func stopspider() {
	close(stop)
}

//页面处理类
type Www79xsComProcessor struct {
}

func NewWww79xsComProcessor() *Www79xsComProcessor {
	return &Www79xsComProcessor{}
}

func (self *Www79xsComProcessor) Process(p *robot.Page) {
	//判断页面是否抓取成功
	if !p.IsSucc() {
		mlog.LogInst().LogError(p.Errormsg())
		return
	}

	//如果callback为空，则说明是入口页面，否则直接执行对应callback
	callback := p.GetRequest().GetCallBack()
	if callback == nil {
		self.mainParse(p)
	} else {
		callback(p)
	}
}

//主页解析
func (self *Www79xsComProcessor) mainParse(p *robot.Page) {
	//开始解析页面
	query := p.GetHtmlParser()
	query.Find(".subnav ul li a").Each(func(i int, s *goquery.Selection) {
		addr, _ := s.Attr("href")
		if addr == girl_url {
			p.AddTargetRequest(initrequest(baseurl+addr, map[string]string{"first": GIRL}, self.urlListParse))
		} else {
			p.AddTargetRequest(initrequest(baseurl+addr, map[string]string{"first": BOY}, self.urlListParse))
		}
	})
}

//分类列表解析
func (self *Www79xsComProcessor) urlListParse(p *robot.Page) {
	//开始解析页面
	//query := p.GetHtmlParser()
	meta := p.GetRequest().GetMeta()
	p.AddField("code", "0")
	p.AddField("first", meta.(map[string]string)["first"])
}

func (self *Www79xsComProcessor) Finish() {
}

//mysql pipeline
type PipelineMySQL struct {
	DB *sql.DB
}

func NewPipelineMySQL(dbinfo string) *PipelineMySQL {
	db, _ := sql.Open("mysql", dbinfo)
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(30)
	if db.Ping() != nil {
		log.Fatalf("connect mysql fail\n")
	}
	return &PipelineMySQL{
		DB: db,
	}
}

func (self *PipelineMySQL) Process(pageitems *robot.PageItems, task robot.Task) {
	//如果code＝“0” 则调用处理一级分类函数
	if code, ok := pageitems.GetItem("code"); ok {
		switch code {
		case "0":
			self.firstProcess(pageitems, task)
		}
	}
}

func (self *PipelineMySQL) firstProcess(pageitems *robot.PageItems, task robot.Task) {
	//接下来存储一级分类和二级分类
}

func main() {
	start_url := baseurl
	//加载配置文件
	settings := loadconf("spider.conf")
	//获取数据库连接信息
	dbinfo, ok := settings["DBINFO"]
	if !ok {
		log.Fatalf("please insert dbinfo in spider.conf")
	}

	//爬虫初始化
	options := robot.SpiderOptions{
		TaskName:      "79xs",
		PageProcesser: NewWww79xsComProcessor(),
		Downloader:    downloader.NewHttpDownloader(),
		Scheduler:     scheduler.NewQueueScheduler(false),
		Pipelines:     []robot.Pipeline{NewPipelineMySQL(dbinfo)},
		//设置资源管理器，资源池容量为10
		ResourceManage: resource.NewSpidersPool(10, nil),
	}

	sp := robot.NewSpider(options)
	//增加根url
	sp.AddRequest(initrequest(start_url, nil, nil))
	go sp.Run()
	<-stop
	sp.Close()
}
