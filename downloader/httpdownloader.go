/*
Author: Aosen
Data: 2015-01-07
Contact: 316052486
Desc: 该http下载器参考自https://github.com/hu17889/go_spider/blob/master/core/downloader/downloader_http.go
*/

package downloader

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/goutils"
	"github.com/aosen/mlog"
	"github.com/aosen/robot"
	"github.com/bitly/go-simplejson"
	"golang.org/x/net/html/charset"
)

// HTTP下载器
// The HttpDownloader download page by package net/http.
// The "html" content is contained in dom parser of package goquery.
// The "json" content is saved.
// The "jsonp" content is modified to json.
// The "text" content will save body plain text only.
// The page result is saved in Page.
type HttpDownloader struct {
	//如果网页中存在不完整的content－type 则默认
	DefaultContentType string
}

func NewHttpDownloader(ct string) *HttpDownloader {
	return &HttpDownloader{
		DefaultContentType: ct,
	}
}

func (self *HttpDownloader) Download(req *robot.Request) *robot.Page {
	var mtype string
	var p = robot.NewPage(req)
	mtype = req.GetResponceType()
	switch mtype {
	case "html":
		return self.downloadHtml(p, req)
	case "json":
		fallthrough
	case "jsonp":
		return self.downloadJson(p, req)
	case "text":
		return self.downloadText(p, req)
	default:
		mlog.LogInst().LogError("error request type:" + mtype)
	}
	return p
}

// choose http GET/method to download
func connectByHttp(p *robot.Page, req *robot.Request) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: req.GetRedirectFunc(),
	}

	httpreq, err := http.NewRequest(req.GetMethod(), req.GetUrl(), strings.NewReader(req.GetPostdata()))
	if header := req.GetHeader(); header != nil {
		httpreq.Header = req.GetHeader()
	}

	if cookies := req.GetCookies(); cookies != nil {
		for i := range cookies {
			httpreq.AddCookie(cookies[i])
		}
	}

	var resp *http.Response
	if resp, err = client.Do(httpreq); err != nil {
		if e, ok := err.(*url.Error); ok && e.Err != nil && e.Err.Error() == "normal" {
			//  normal
		} else {
			mlog.LogInst().LogError(err.Error())
			p.SetStatus(true, err.Error())
			//fmt.Printf("client do error %v \r\n", err)
			return nil, err
		}
	}

	return resp, nil
}

// choose a proxy server to excute http GET/method to download
func connectByHttpProxy(p *robot.Page, in_req *robot.Request) (*http.Response, error) {
	request, _ := http.NewRequest("GET", in_req.GetUrl(), nil)
	proxy, err := url.Parse(in_req.GetProxyHost())
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

// Charset auto determine. Use golang.org/x/net/html/charset. Get page body and change it to utf-8
// 自动转码
func (self *HttpDownloader) changeCharsetEncodingAuto(contentTypeStr string, sor io.ReadCloser) string {
	if len(strings.Split(contentTypeStr, " ")) < 2 {
		contentTypeStr = self.DefaultContentType
	}
	destReader, err := charset.NewReader(sor, contentTypeStr)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		destReader = sor
	}
	var sorbody []byte
	sorbody, err = ioutil.ReadAll(destReader)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
	}
	bodystr := string(sorbody)
	return bodystr
}

func (self *HttpDownloader) changeCharsetEncodingAutoGzipSupport(contentTypeStr string, sor io.ReadCloser) string {
	var err error
	gzipReader, err := gzip.NewReader(sor)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		return ""
	}
	defer gzipReader.Close()
	return self.changeCharsetEncodingAuto(contentTypeStr, gzipReader)
}

// Download file and change the charset of page charset.
func (self *HttpDownloader) downloadFile(p *robot.Page, req *robot.Request) (*robot.Page, string) {
	var err error
	var urlstr string
	if urlstr = req.GetUrl(); len(urlstr) == 0 {
		mlog.LogInst().LogError("url is empty")
		p.SetStatus(true, "url is empty")
		return p, ""
	}

	var resp *http.Response

	if proxystr := req.GetProxyHost(); len(proxystr) != 0 {
		//using http proxy
		//fmt.Print("HttpProxy Enter ",proxystr,"\n")
		resp, err = connectByHttpProxy(p, req)
	} else {
		//normal http download
		//fmt.Print("Http Normal Enter \n",proxystr,"\n")
		resp, err = connectByHttp(p, req)
	}

	if err != nil {
		return p, ""
	}

	p.SetHeader(resp.Header)
	p.SetCookies(resp.Cookies())

	// get converter to utf-8
	var bodyStr string
	if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyStr = self.changeCharsetEncodingAutoGzipSupport(resp.Header.Get("Content-Type"), resp.Body)
	} else {
		bodyStr = self.changeCharsetEncodingAuto(resp.Header.Get("Content-Type"), resp.Body)
	}
	//fmt.Printf("utf-8 body %v \r\n", bodyStr)
	defer resp.Body.Close()
	return p, bodyStr
}

func (self *HttpDownloader) downloadHtml(p *robot.Page, req *robot.Request) *robot.Page {
	var err error
	p, destbody := self.downloadFile(p, req)
	//fmt.Printf("Destbody %v \r\n", destbody)
	if !p.IsSucc() {
		//fmt.Print("Page error \r\n")
		return p
	}
	bodyReader := bytes.NewReader([]byte(destbody))

	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(bodyReader); err != nil {
		mlog.LogInst().LogError(err.Error())
		p.SetStatus(true, err.Error())
		return p
	}

	var body string
	if body, err = doc.Html(); err != nil {
		mlog.LogInst().LogError(err.Error())
		p.SetStatus(true, err.Error())
		return p
	}

	p.SetBodyStr(body).SetHtmlParser(doc).SetStatus(false, "")

	return p
}

func (self *HttpDownloader) downloadJson(p *robot.Page, req *robot.Request) *robot.Page {
	var err error
	p, destbody := self.downloadFile(p, req)
	if !p.IsSucc() {
		return p
	}

	var body []byte
	body = []byte(destbody)
	mtype := req.GetResponceType()
	if mtype == "jsonp" {
		tmpstr := goutils.JsonpToJson(destbody)
		body = []byte(tmpstr)
	}

	var r *simplejson.Json
	if r, err = simplejson.NewJson(body); err != nil {
		mlog.LogInst().LogError(string(body) + "\t" + err.Error())
		p.SetStatus(true, err.Error())
		return p
	}

	// json result
	p.SetBodyStr(string(body)).SetJson(r).SetStatus(false, "")

	return p
}

func (self *HttpDownloader) downloadText(p *robot.Page, req *robot.Request) *robot.Page {
	p, destbody := self.downloadFile(p, req)
	if !p.IsSucc() {
		return p
	}

	p.SetBodyStr(destbody).SetStatus(false, "")
	return p
}
