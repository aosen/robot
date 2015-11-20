package robot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"testing"
)

func TestResourceManage(t *testing.T) {
	var mc *ResourceManageChan
	mc = NewResourceManageChan(1)
	mc.GetOne()
	println("incr")
	mc.FreeOne()
	println("decr")
	mc.GetOne()
	println("incr")
}

func Test_QueueScheduler(t *testing.T) {
	var r *Request
	r = NewRequest("http://baidu.com", "html", "", "GET", "", nil, nil, nil, nil)
	fmt.Printf("%v\n", r)

	var s *QueueScheduler
	s = NewQueueScheduler(false)

	s.Push(r)
	var count int = s.Count()
	if count != 1 {
		t.Error("count error")
	}
	fmt.Println(count)

	var r1 *Request
	r1 = s.Poll()
	if r1 == nil {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)

	// remove duplicate
	s = NewQueueScheduler(true)

	r2 := NewRequest("http://qq.com", "html", "", "GET", "", nil, nil, nil, nil)
	s.Push(r)
	s.Push(r2)
	s.Push(r)
	count = s.Count()
	if count != 2 {
		t.Error("count error")
	}
	fmt.Println(count)

	r1 = s.Poll()
	if r1 == nil {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)
	r1 = s.Poll()
	if r1 == nil {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)
}

func Test_DownloadHtml(t *testing.T) {
	//return
	//request := request.NewRequest("http://live.sina.com.cn/zt/api/l/get/finance/globalnews1/index.htm?format=json&callback=t13975294&id=23521&pagesize=45&dire=f&dpc=1")
	var req *Request
	req = NewRequest("http://live.sina.com.cn/zt/l/v/finance/globalnews1/", "html", "", "GET", "", nil, nil, nil, nil)

	var dl Downloader
	dl = NewHttpDownloader()

	var p *Page
	p = dl.Download(req)

	var doc *goquery.Document
	doc = p.GetHtmlParser()
	//fmt.Println(doc)
	//body := p.GetBodyStr()
	//fmt.Println(body)

	var s *goquery.Selection
	s = doc.Find("body")
	if s.Length() < 1 {
		t.Error("html parse failed!")
	}

	/*
	   doc, err := goquery.NewDocument("http://live.sina.com.cn/zt/l/v/finance/globalnews1/")
	   if err != nil {
	       fmt.Printf("%v",err)
	   }
	   s := doc.Find("meta");
	   fmt.Println(s.Length())

	   resp, err := http.Get("http://live.sina.com.cn/zt/l/v/finance/globalnews1/")
	   if err != nil {
	       fmt.Printf("%v",err)
	   }
	   defer resp.Body.Close()
	   doc, err = goquery.NewDocumentFromReader(resp.Body)
	   s = doc.Find("meta");
	   fmt.Println(s.Length())
	*/
}

func Test_DownloadJson(t *testing.T) {
	//return
	var req *Request
	req = NewRequest("http://live.sina.com.cn/zt/api/l/get/finance/globalnews1/index.htm?format=json&id=23521&pagesize=4&dire=f&dpc=1", "json", "", "GET", "", nil, nil, nil, nil)

	var dl Downloader
	dl = NewHttpDownloader()

	var p *Page
	p = dl.Download(req)

	var jsonMap interface{}
	jsonMap = p.GetJson()
	fmt.Printf("%v", jsonMap)

	//fmt.Println(doc)
	//body := p.GetBodyStr()
	//fmt.Println(body)

}

func Test_CharSetChange(t *testing.T) {
	var req *Request
	//req = request.NewRequest("http://stock.finance.sina.com.cn/usstock/api/jsonp.php/t/US_CategoryService.getList?page=1&num=60", "jsonp")
	req = NewRequest("http://soft.chinabyte.com/416/13164916.shtml", "html", "", "GET", "", nil, nil, nil, nil)

	var dl Downloader
	dl = NewHttpDownloader()

	var p *Page
	p = dl.Download(req)

	//hp := p.GetHtmlParser()
	//fmt.Printf("%v", jsonMap)

	//fmt.Println(doc)
	p.GetBodyStr()
	body := p.GetBodyStr()
	fmt.Println(body)

}
