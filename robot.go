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
func (this *Request) AddHeaderFile(headerFile string) *Request {
	_, err := os.Stat(headerFile)
	if err != nil {
		return this
	}
	h := readHeaderFromFile(headerFile)
	this.Header = h
	return this
}

// @host  http://localhost:8765/
func (this *Request) AddProxyHost(host string) *Request {
	this.ProxyHost = host
	return this
}

func (this *Request) GetUrl() string {
	return this.Url
}

func (this *Request) GetUrlTag() string {
	return this.Urltag
}

func (this *Request) GetMethod() string {
	return this.Method
}

func (this *Request) GetPostdata() string {
	return this.Postdata
}

func (this *Request) GetHeader() http.Header {
	return this.Header
}

func (this *Request) GetCookies() []*http.Cookie {
	return this.Cookies
}

func (this *Request) GetProxyHost() string {
	return this.ProxyHost
}

func (this *Request) GetResponceType() string {
	return this.RespType
}

func (this *Request) GetRedirectFunc() func(req *http.Request, via []*http.Request) error {
	return this.checkRedirect
}

func (this *Request) GetMeta() interface{} {
	return this.Meta
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
func (this *PageItems) GetRequest() *Request {
	return this.req
}

// AddItem saves a KV result into PageItems.
func (this *PageItems) AddItem(key string, item string) {
	this.items[key] = item
}

// GetItem returns value of the key.
func (this *PageItems) GetItem(key string) (string, bool) {
	t, ok := this.items[key]
	return t, ok
}

// GetAll returns all the KVs result.
func (this *PageItems) GetAll() map[string]string {
	return this.items
}

// SetSkip set skip true to make this page not to be processed by Pipeline.
func (this *PageItems) SetSkip(skip bool) *PageItems {
	this.skip = skip
	return this
}

// GetSkip returns skip label.
func (this *PageItems) GetSkip() bool {
	return this.skip
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
func (this *Page) SetStatus(isfail bool, errormsg string) {
	this.isfail = isfail
	this.errormsg = errormsg
}

// SetHeader save the header of http responce
func (this *Page) SetHeader(header http.Header) {
	this.header = header
}

// GetHeader returns the header of http responce
func (this *Page) GetHeader() http.Header {
	return this.header
}

// SetHeader save the cookies of http responce
func (this *Page) SetCookies(cookies []*http.Cookie) {
	this.cookies = cookies
}

// GetHeader returns the cookies of http responce
func (this *Page) GetCookies() []*http.Cookie {
	return this.cookies
}

// IsSucc test whether download process success or not.
func (this *Page) IsSucc() bool {
	return !this.isfail
}

// Errormsg show the download error message.
func (this *Page) Errormsg() string {
	return this.errormsg
}

// AddField saves KV string pair to PageItems preparing for Pipeline
func (this *Page) AddField(key string, value string) {
	this.pItems.AddItem(key, value)
}

// GetPageItems returns PageItems object that record KV pair parsed in PageProcesser.
func (this *Page) GetPageItems() *PageItems {
	return this.pItems
}

// SetSkip set label "skip" of PageItems.
// PageItems will not be saved in Pipeline wher skip is set true
func (this *Page) SetSkip(skip bool) {
	this.pItems.SetSkip(skip)
}

// GetSkip returns skip label of PageItems.
func (this *Page) GetSkip() bool {
	return this.pItems.GetSkip()
}

// SetRequest saves request oject of this page.
func (this *Page) SetRequest(r *Request) *Page {
	this.req = r
	return this
}

// GetRequest returns request oject of this page.
func (this *Page) GetRequest() *Request {
	return this.req
}

// GetUrlTag returns name of url.
func (this *Page) GetUrlTag() string {
	return this.req.GetUrlTag()
}

// AddTargetRequest adds one new Request waitting for crawl.
func (this *Page) AddTargetRequest(url string, respType string) *Page {
	this.targetRequests = append(this.targetRequests, NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil))
	return this
}

// AddTargetRequests adds new Requests waitting for crawl.
func (this *Page) AddTargetRequests(urls []string, respType string) *Page {
	for _, url := range urls {
		this.AddTargetRequest(url, respType)
	}
	return this
}

// AddTargetRequestWithProxy adds one new Request waitting for crawl.
func (this *Page) AddTargetRequestWithProxy(url string, respType string, proxyHost string) *Page {

	this.targetRequests = append(this.targetRequests, NewRequestWithProxy(url, respType, "", "GET", "", nil, nil, proxyHost, nil, nil))
	return this
}

// AddTargetRequestsWithProxy adds new Requests waitting for crawl.
func (this *Page) AddTargetRequestsWithProxy(urls []string, respType string, proxyHost string) *Page {
	for _, url := range urls {
		this.AddTargetRequestWithProxy(url, respType, proxyHost)
	}
	return this
}

// AddTargetRequest adds one new Request with header file for waitting for crawl.
func (this *Page) AddTargetRequestWithHeaderFile(url string, respType string, headerFile string) *Page {
	this.targetRequests = append(this.targetRequests, NewRequestWithHeaderFile(url, respType, headerFile))
	return this
}

// AddTargetRequest adds one new Request waitting for crawl.
// The respType is "html" or "json" or "jsonp" or "text".
// The urltag is name for marking url and distinguish different urls in PageProcesser and Pipeline.
// The method is POST or GET.
// The postdata is http body string.
// The header is http header.
// The cookies is http cookies.
func (this *Page) AddTargetRequestWithParams(req *Request) *Page {
	this.targetRequests = append(this.targetRequests, req)
	return this
}

// AddTargetRequests adds new Requests waitting for crawl.
func (this *Page) AddTargetRequestsWithParams(reqs []*Request) *Page {
	for _, req := range reqs {
		this.AddTargetRequestWithParams(req)
	}
	return this
}

// GetTargetRequests returns the target requests that will put into Scheduler
func (this *Page) GetTargetRequests() []*Request {
	return this.targetRequests
}

// SetBodyStr saves plain string crawled in Page.
func (this *Page) SetBodyStr(body string) *Page {
	this.body = body
	return this
}

// GetBodyStr returns plain string crawled.
func (this *Page) GetBodyStr() string {
	return this.body
}

// SetHtmlParser saves goquery object binded to target crawl result.
func (this *Page) SetHtmlParser(doc *goquery.Document) *Page {
	this.docParser = doc
	return this
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (this *Page) GetHtmlParser() *goquery.Document {
	return this.docParser
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (this *Page) ResetHtmlParser() *goquery.Document {
	r := strings.NewReader(this.body)
	var err error
	this.docParser, err = goquery.NewDocumentFromReader(r)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		panic(err.Error())
	}
	return this.docParser
}

// SetJson saves json result.
func (this *Page) SetJson(js *simplejson.Json) *Page {
	this.jsonMap = js
	return this
}

// SetJson returns json result.
func (this *Page) GetJson() *simplejson.Json {
	return this.jsonMap
}
