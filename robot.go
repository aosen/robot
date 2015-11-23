package robot

/*
对象：
调度器(Scheduler)
下载器(HttpDownloader)
项目管道(Pipeline)
*/

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/mlog"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

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

	// Redirect function for downloader used in http.Client
	// If CheckRedirect returns an error, the Client's Get
	// method returns both the previous Response.
	// If CheckRedirect returns error.New("normal"), the error process after client.Do will ignore the error.
	checkRedirect func(req *http.Request, via []*http.Request) error

	Meta interface{}
}

// NewRequest returns initialized Request object.
// The respType is json, jsonp, html, text
/*
func NewRequestSimple(url string, respType string, urltag string) *Request {
    return &Request{url:url, respType:respType}
}
*/

func NewRequest(url string, respType string, urltag string, method string,
	postdata string, header http.Header, cookies []*http.Cookie,
	checkRedirect func(req *http.Request, via []*http.Request) error,
	meta interface{}) *Request {
	return &Request{url, respType, method, postdata, urltag, header, cookies, "", checkRedirect, meta}
}

func NewRequestWithProxy(url string, respType string, urltag string, method string,
	postdata string, header http.Header, cookies []*http.Cookie, proxyHost string,
	checkRedirect func(req *http.Request, via []*http.Request) error,
	meta interface{}) *Request {
	return &Request{url, respType, method, postdata, urltag, header, cookies, proxyHost, checkRedirect, meta}
}

func NewRequestWithHeaderFile(url string, respType string, headerFile string) *Request {
	_, err := os.Stat(headerFile)
	if err != nil {
		//file is not exist , using default mode
		return NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
	}

	h := readHeaderFromFile(headerFile)

	return NewRequest(url, respType, "", "GET", "", h, nil, nil, nil)
}

func readHeaderFromFile(headerFile string) http.Header {
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
	h := readHeaderFromFile(headerFile)
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

func (self *Request) GetRedirectFunc() func(req *http.Request, via []*http.Request) error {
	return self.checkRedirect
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
	return &Page{pItems: NewPageItems(req), req: req}
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
func (self *Page) AddTargetRequest(url string, respType string) *Page {
	self.targetRequests = append(self.targetRequests, NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil))
	return self
}

// AddTargetRequests adds new Requests waitting for crawl.
func (self *Page) AddTargetRequests(urls []string, respType string) *Page {
	for _, url := range urls {
		self.AddTargetRequest(url, respType)
	}
	return self
}

// AddTargetRequestWithProxy adds one new Request waitting for crawl.
func (self *Page) AddTargetRequestWithProxy(url string, respType string, proxyHost string) *Page {

	self.targetRequests = append(self.targetRequests, NewRequestWithProxy(url, respType, "", "GET", "", nil, nil, proxyHost, nil, nil))
	return self
}

// AddTargetRequestsWithProxy adds new Requests waitting for crawl.
func (self *Page) AddTargetRequestsWithProxy(urls []string, respType string, proxyHost string) *Page {
	for _, url := range urls {
		self.AddTargetRequestWithProxy(url, respType, proxyHost)
	}
	return self
}

// AddTargetRequest adds one new Request with header file for waitting for crawl.
func (self *Page) AddTargetRequestWithHeaderFile(url string, respType string, headerFile string) *Page {
	self.targetRequests = append(self.targetRequests, NewRequestWithHeaderFile(url, respType, headerFile))
	return self
}

// AddTargetRequest adds one new Request waitting for crawl.
// The respType is "html" or "json" or "jsonp" or "text".
// The urltag is name for marking url and distinguish different urls in PageProcesser and Pipeline.
// The method is POST or GET.
// The postdata is http body string.
// The header is http header.
// The cookies is http cookies.
func (self *Page) AddTargetRequestWithParams(req *Request) *Page {
	self.targetRequests = append(self.targetRequests, req)
	return self
}

// AddTargetRequests adds new Requests waitting for crawl.
func (self *Page) AddTargetRequestsWithParams(reqs []*Request) *Page {
	for _, req := range reqs {
		self.AddTargetRequestWithParams(req)
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
