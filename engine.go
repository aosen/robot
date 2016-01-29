/*
Author: Aosen
Data: 2016-01-18
QQ: 316052486
Desc: 爬虫引擎
*/

package robot

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/mlog"
	"github.com/bitly/go-simplejson"
)

type Spider struct {
	taskname string

	pageProcesser PageProcesser

	downloader Downloader

	scheduler Scheduler

	pipelines []Pipeline

	rm               ResourceManage
	exitWhenComplete bool

	// Sleeptype can be fixed or rand.
	startSleeptime uint
	endSleeptime   uint
	sleeptype      string
}

//爬虫设置选项
type SpiderOptions struct {
	//任务名称
	TaskName string
	//页面处理接口实现
	PageProcesser PageProcesser
	//下载器接口实现
	Downloader Downloader
	//调度器接口实现
	Scheduler Scheduler
	//Pipeline的接口实现，直接将一系列pipeline的实现对象放入这个列表
	Pipelines []Pipeline
	//资源管理接口实现
	ResourceManage ResourceManage
	//最大协程数,用于协程池
	MaxGoroutineNum uint
}

/*
2016-01-07
创建爬虫项目，一切从这个开始，首选需要你添加爬虫的各种选项参数，包括用哪种下载器，哪种调度器，哪种资源管理器，哪种pipeline，及页面处理器
当然，我们也为你准备了一系列写好的类，给你进行参考和使用，你可以到对应的文件夹中去寻找
*/
func NewSpider(options SpiderOptions) *Spider {
	//开启日志
	mlog.StraceInst().Open()
	sp := &Spider{
		taskname:      options.TaskName,
		pageProcesser: options.PageProcesser,
		downloader:    options.Downloader,
		scheduler:     options.Scheduler,
		pipelines:     options.Pipelines,
		rm:            options.ResourceManage,
	}
	// init filelog.
	sp.CloseFileLog()
	sp.sleeptype = "fixed"
	sp.startSleeptime = 0

	if sp.taskname == "" {
		sp.taskname = "robot"
	}
	if sp.scheduler == nil {
		log.Fatal("Please choose the need to use the Scheduler.")
	}
	if sp.downloader == nil {
		log.Fatal("Please choose the need to use the Downloader")
	}
	if sp.pageProcesser == nil {
		log.Fatal("Please choose the need to use the PageProcesser")
	}
	if sp.rm == nil {
		log.Fatal("Please choose the need to use the ResourceManage")
	}
	mlog.StraceInst().Println(sp.taskname + " " + "start")
	return sp
}

// Deal with one url and return the PageItems.
func (self *Spider) Get(req *Request) *PageItems {
	return self.GetByRequest(req)
}

// Deal with one url and return the PageItems with other setting.
func (self *Spider) GetByRequest(req *Request) *PageItems {
	var reqs []*Request
	reqs = append(reqs, req)
	items := self.GetAllByRequest(reqs)
	if len(items) != 0 {
		return items[0]
	}
	return nil
}

// Deal with several urls and return the PageItems slice
func (self *Spider) GetAllByRequest(reqs []*Request) []*PageItems {
	// push url
	for _, req := range reqs {
		//req := request.NewRequest(u, respType, urltag, method, postdata, header, cookies)
		self.AddRequest(req)
	}

	pip := NewCollectPipelinePageItems()
	self.AddPipeline(pip)

	self.Run()

	return pip.GetCollected()
}

// add Request to Schedule
func (self *Spider) AddRequest(req *Request) *Spider {
	if req == nil {
		mlog.LogInst().LogError("request is nil")
		return self
	} else if req.GetUrl() == "" {
		mlog.LogInst().LogError("request is empty")
		return self
	}
	self.scheduler.Push(req)
	return self
}

//
func (self *Spider) AddRequests(reqs []*Request) *Spider {
	for _, req := range reqs {
		self.AddRequest(req)
	}
	return self
}

func (self *Spider) AddPipeline(p Pipeline) *Spider {
	self.pipelines = append(self.pipelines, p)
	return self
}

// Deal with several urls and return the PageItems slice.
func (self *Spider) GetAll(reqs []*Request) []*PageItems {
	for _, req := range reqs {
		self.AddRequest(req)
	}

	pip := NewCollectPipelinePageItems()
	self.AddPipeline(pip)

	self.Run()

	return pip.GetCollected()
}

func (self *Spider) Taskname() string {
	return self.taskname
}

// The CloseFileLog close file log.
func (self *Spider) CloseFileLog() *Spider {
	mlog.InitFilelog(false, "")
	return self
}

func (self *Spider) SetDownloader(d Downloader) *Spider {
	self.downloader = d
	return self
}

func (self *Spider) GetScheduler() Scheduler {
	return self.scheduler
}

func (self *Spider) SetScheduler(s Scheduler) *Spider {
	self.scheduler = s
	return self
}

func (self *Spider) Run() {
	//不断向爬虫池添加任务
	go func() {
		for {
			req := self.scheduler.Poll()

			// rm is not atomic
			if self.rm.Has() == 0 && req == nil && self.exitWhenComplete {
				self.pageProcesser.Finish()
				mlog.StraceInst().Println("Grab complete !!!")
				break
			} else if req == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			self.rm.AddTask(func(req *Request) {
				log.Println("start crawl : " + req.GetUrl() + " urls:" + strconv.Itoa(self.scheduler.Count()))
				self.pageProcess(req)
			}, req)
		}
		//关闭爬虫
		self.Close()
		//释放爬虫池
		self.rm.Free()
	}()
	//爬虫池开始执行任务
	self.rm.Start()
}

func (self *Spider) Close() {
	//self.SetScheduler(NewQueueScheduler(false))
	//self.SetDownloader(NewHttpDownloader())
	self.pipelines = make([]Pipeline, 0)
	self.exitWhenComplete = true
	mlog.StraceInst().Println("stop crawl")
}

// core processer
func (self *Spider) pageProcess(req *Request) {
	var p *Page

	defer func() {
		if err := recover(); err != nil { // do not affect other
			if strerr, ok := err.(string); ok {
				mlog.LogInst().LogError(strerr)
			} else {
				mlog.LogInst().LogError("pageProcess error")
			}
		}
	}()

	// download page
	for i := 0; i < 3; i++ {
		self.sleep()
		p = self.downloader.Download(req)
		if p.IsSucc() { // if fail retry 3 times
			break
		}

	}

	if !p.IsSucc() { // if fail do not need process
		return
	}

	self.pageProcesser.Process(p)
	//将targetRequests中的所有请求列表放入调度队列
	for _, req := range p.GetTargetRequests() {
		self.AddRequest(req)
	}

	// output
	if !p.GetSkip() {
		for _, pip := range self.pipelines {
			//fmt.Println("%v",p.GetPageItems().GetAll())
			pip.Process(p.GetPageItems(), self)
		}
	}
}

func (self *Spider) sleep() {
	if self.sleeptype == "fixed" {
		time.Sleep(time.Duration(self.startSleeptime) * time.Millisecond)
	} else if self.sleeptype == "rand" {
		sleeptime := rand.Intn(int(self.endSleeptime-self.startSleeptime)) + int(self.startSleeptime)
		time.Sleep(time.Duration(sleeptime) * time.Millisecond)
	}
}

/*
用户函数
*/

func (self *Spider) GetDownloader() Downloader {
	return self.downloader
}

// If exit when each crawl task is done.
// If you want to keep spider in memory all the time and add url from outside, you can set it true.
func (self *Spider) SetExitWhenComplete(e bool) *Spider {
	self.exitWhenComplete = e
	return self
}

func (self *Spider) GetExitWhenComplete() bool {
	return self.exitWhenComplete
}

// The OpenFileLog initialize the log path and open log.
// If log is opened, error info or other useful info in spider will be logged in file of the filepath.
// Log command is mlog.LogInst().LogError("info") or mlog.LogInst().LogInfo("info").
// Spider's default log is closed.
// The filepath is absolute path.
func (self *Spider) OpenFileLog(filePath string) *Spider {
	mlog.InitFilelog(true, filePath)
	return self
}

// OpenFileLogDefault open file log with default file path like "WD/log/log.2014-9-1".
func (self *Spider) OpenFileLogDefault() *Spider {
	mlog.InitFilelog(true, "")
	return self
}

// The OpenStrace open strace that output progress info on the screen.
// Spider's default strace is opened.
func (self *Spider) OpenStrace() *Spider {
	mlog.StraceInst().Open()
	return self
}

// The CloseStrace close strace.
func (self *Spider) CloseStrace() *Spider {
	mlog.StraceInst().Close()
	return self
}

// The SetSleepTime set sleep time after each crawl task.
// The unit is millisecond.
// If sleeptype is "fixed", the s is the sleep time and e is useless.
// If sleeptype is "rand", the sleep time is rand between s and e.
func (self *Spider) SetSleepTime(sleeptype string, s uint, e uint) *Spider {
	self.sleeptype = sleeptype
	self.startSleeptime = s
	self.endSleeptime = e
	if self.sleeptype == "rand" && self.startSleeptime >= self.endSleeptime {
		panic("startSleeptime must smaller than endSleeptime")
	}
	return self
}

// Request represents object waiting for being crawled.
type Request struct {
	Url string

	// Responce type: html json jsonp text
	RespType string

	// GET POST
	Method string

	// POST data
	Postdata string

	// name for marking url and distinguish different urls in PageProcesser and Pipeline
	Urltag string

	// http header
	Header http.Header

	// http cookies
	Cookies []*http.Cookie

	//proxy host   example='localhost:80'
	ProxyHost string

	Meta interface{}
}

func NewRequest(req *Request) *Request {
	//主要做默认值设置与错误检测
	if req.Url == "" {
		mlog.LogInst().LogError("request url is nil")
	}
	if req.Method == "" || req.Method != "GET" || req.Method != "POST" || req.Method != "HEAD" || req.Method != "OPTIONS" || req.Method != "PUT" || req.Method != "DELETE" {
		req.Method = "GET"
	}
	if req.RespType == "" || req.RespType != "html" || req.RespType != "json" || req.RespType != "jsonp" || req.RespType != "text" {
		req.RespType = "html"
	}
	return req
}

func ReadHeaderFromFile(headerFile string) http.Header {
	//read file , parse the header and cookies
	b, err := ioutil.ReadFile(headerFile)
	if err != nil {
		//make be:  share access error
		mlog.LogInst().LogError(err.Error())
		return nil
	}
	js, _ := simplejson.NewJson(b)
	//constructed to header

	h := make(http.Header)
	h.Add("User-Agent", js.Get("User-Agent").MustString())
	h.Add("Referer", js.Get("Referer").MustString())
	h.Add("Cookie", js.Get("Cookie").MustString())
	h.Add("Cache-Control", "max-age=0")
	h.Add("Connection", "keep-alive")
	return h
}

//point to a json file
/* xxx.json
{
	"User-Agent":"curl/7.19.3 (i386-pc-win32) libcurl/7.19.3 OpenSSL/1.0.0d",
	"Referer":"http://weixin.sogou.com/gzh?openid=oIWsFt6Sb7aZmuI98AU7IXlbjJps",
	"Cookie":""
}
*/
func (self *Request) AddHeaderFile(headerFile string) *Request {
	_, err := os.Stat(headerFile)
	if err != nil {
		return self
	}
	h := ReadHeaderFromFile(headerFile)
	self.Header = h
	return self
}

// @host  http://localhost:8765/
func (self *Request) AddProxyHost(host string) *Request {
	self.ProxyHost = host
	return self
}

func (self *Request) GetUrl() string {
	return self.Url
}

//获取URL路径
//http://www.79xs.com/Html/Book/147/147144/Index.html
//返回http://www.79xs.com/Html/Book/147/147144/
func (self *Request) GetBaseUrl() string {
	sl := strings.SplitAfter(self.Url, "/")
	if len(sl) == 0 {
		return ""
	} else {
		return strings.Join(sl[:len(sl)-1], "")
	}
}

func (self *Request) GetUrlTag() string {
	return self.Urltag
}

func (self *Request) GetMethod() string {
	return self.Method
}

func (self *Request) GetPostdata() string {
	return self.Postdata
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) GetCookies() []*http.Cookie {
	return self.Cookies
}

func (self *Request) GetProxyHost() string {
	return self.ProxyHost
}

func (self *Request) GetResponceType() string {
	return self.RespType
}

func (self *Request) GetMeta() interface{} {
	return self.Meta
}

// PageItems represents an entity save result parsed by PageProcesser and will be output at last.
//保存解析后结果
type PageItems struct {

	// The req is Request object that contains the parsed result, which saved in PageItems.
	req *Request

	// The items is the container of parsed result.
	items map[string]string

	// The skip represents whether send ResultItems to scheduler or not.
	skip bool
}

// NewPageItems returns initialized PageItems object.
// 返回一个初始化的pageitems
func NewPageItems(req *Request) *PageItems {
	items := make(map[string]string)
	return &PageItems{req: req, items: items, skip: false}
}

// GetRequest returns request of PageItems
func (self *PageItems) GetRequest() *Request {
	return self.req
}

// AddItem saves a KV result into PageItems.
func (self *PageItems) AddItem(key string, item string) {
	self.items[key] = item
}

// GetItem returns value of the key.
func (self *PageItems) GetItem(key string) (string, bool) {
	t, ok := self.items[key]
	return t, ok
}

// GetAll returns all the KVs result.
func (self *PageItems) GetAll() map[string]string {
	return self.items
}

// SetSkip set skip true to make self page not to be processed by Pipeline.
func (self *PageItems) SetSkip(skip bool) *PageItems {
	self.skip = skip
	return self
}

// GetSkip returns skip label.
func (self *PageItems) GetSkip() bool {
	return self.skip
}

// Page represents an entity be crawled.
type Page struct {
	// The isfail is true when crawl process is failed and errormsg is the fail resean.
	isfail   bool
	errormsg string

	// The request is crawled by spider that contains url and relevent information.
	req *Request

	// The body is plain text of crawl result.
	body string

	header  http.Header
	cookies []*http.Cookie

	// The docParser is a pointer of goquery boject that contains html result.
	docParser *goquery.Document

	// The jsonMap is the json result.
	jsonMap *simplejson.Json

	// The pItems is object for save Key-Values in PageProcesser.
	// And pItems is output in Pipline.
	pItems *PageItems

	// The targetRequests is requests to put into Scheduler.
	targetRequests []*Request
}

// NewPage returns initialized Page object.
func NewPage(req *Request) *Page {
	return &Page{
		pItems: NewPageItems(req),
		req:    req,
	}
}

// SetStatus save status info about download process.
func (self *Page) SetStatus(isfail bool, errormsg string) {
	self.isfail = isfail
	self.errormsg = errormsg
}

// SetHeader save the header of http responce
func (self *Page) SetHeader(header http.Header) {
	self.header = header
}

// GetHeader returns the header of http responce
func (self *Page) GetHeader() http.Header {
	return self.header
}

// SetHeader save the cookies of http responce
func (self *Page) SetCookies(cookies []*http.Cookie) {
	self.cookies = cookies
}

// GetHeader returns the cookies of http responce
func (self *Page) GetCookies() []*http.Cookie {
	return self.cookies
}

// IsSucc test whether download process success or not.
func (self *Page) IsSucc() bool {
	return !self.isfail
}

// Errormsg show the download error message.
func (self *Page) Errormsg() string {
	return self.errormsg
}

// AddField saves KV string pair to PageItems preparing for Pipeline
func (self *Page) AddField(key string, value string) {
	self.pItems.AddItem(key, value)
}

// GetPageItems returns PageItems object that record KV pair parsed in PageProcesser.
func (self *Page) GetPageItems() *PageItems {
	return self.pItems
}

// SetSkip set label "skip" of PageItems.
// PageItems will not be saved in Pipeline wher skip is set true
func (self *Page) SetSkip(skip bool) {
	self.pItems.SetSkip(skip)
}

// GetSkip returns skip label of PageItems.
func (self *Page) GetSkip() bool {
	return self.pItems.GetSkip()
}

// SetRequest saves request oject of self page.
func (self *Page) SetRequest(r *Request) *Page {
	self.req = r
	return self
}

// GetRequest returns request oject of self page.
func (self *Page) GetRequest() *Request {
	return self.req
}

// GetUrlTag returns name of url.
func (self *Page) GetUrlTag() string {
	return self.req.GetUrlTag()
}

// AddTargetRequest adds one new Request waitting for crawl.
func (self *Page) AddTargetRequest(req *Request) *Page {
	self.targetRequests = append(self.targetRequests, NewRequest(req))
	return self
}

// AddTargetRequests adds new Requests waitting for crawl.
func (self *Page) AddTargetRequests(reqs []*Request) *Page {
	for _, req := range reqs {
		self.AddTargetRequest(NewRequest(req))
	}
	return self
}

// GetTargetRequests returns the target requests that will put into Scheduler
func (self *Page) GetTargetRequests() []*Request {
	return self.targetRequests
}

// SetBodyStr saves plain string crawled in Page.
func (self *Page) SetBodyStr(body string) *Page {
	self.body = body
	return self
}

// GetBodyStr returns plain string crawled.
func (self *Page) GetBodyStr() string {
	return self.body
}

// SetHtmlParser saves goquery object binded to target crawl result.
func (self *Page) SetHtmlParser(doc *goquery.Document) *Page {
	self.docParser = doc
	return self
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Page) GetHtmlParser() *goquery.Document {
	return self.docParser
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Page) ResetHtmlParser() *goquery.Document {
	r := strings.NewReader(self.body)
	var err error
	self.docParser, err = goquery.NewDocumentFromReader(r)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		panic(err.Error())
	}
	return self.docParser
}

// SetJson saves json result.
func (self *Page) SetJson(js *simplejson.Json) *Page {
	self.jsonMap = js
	return self
}

// SetJson returns json result.
func (self *Page) GetJson() *simplejson.Json {
	return self.jsonMap
}

type CollectPipelinePageItems struct {
	collector []*PageItems
}

func NewCollectPipelinePageItems() *CollectPipelinePageItems {
	collector := make([]*PageItems, 0)
	return &CollectPipelinePageItems{collector: collector}
}

func (self *CollectPipelinePageItems) Process(items *PageItems, t Task) {
	self.collector = append(self.collector, items)
}

func (self *CollectPipelinePageItems) GetCollected() []*PageItems {
	return self.collector
}
