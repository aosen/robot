package downloader

import (
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/robot"
)

func Test_DownloadHtml(t *testing.T) {
	//return
	//request := request.NewRequest("http://live.sina.com.cn/zt/api/l/get/finance/globalnews1/index.htm?format=json&callback=t13975294&id=23521&pagesize=45&dire=f&dpc=1")
	var req *robot.Request
	req = robot.NewRequest("http://live.sina.com.cn/zt/l/v/finance/globalnews1/", "html", "", "GET", "", nil, nil, nil, nil)

	var dl robot.Downloader
	dl = NewHttpDownloader()

	var p *robot.Page
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
	var req *robot.Request
	req = robot.NewRequest("http://live.sina.com.cn/zt/api/l/get/finance/globalnews1/index.htm?format=json&id=23521&pagesize=4&dire=f&dpc=1", "json", "", "GET", "", nil, nil, nil, nil)

	var dl robot.Downloader
	dl = NewHttpDownloader()

	var p *robot.Page
	p = dl.Download(req)

	var jsonMap interface{}
	jsonMap = p.GetJson()
	fmt.Printf("%v", jsonMap)

	//fmt.Println(doc)
	//body := p.GetBodyStr()
	//fmt.Println(body)

}

func Test_CharSetChange(t *testing.T) {
	var req *robot.Request
	//req = request.NewRequest("http://stock.finance.sina.com.cn/usstock/api/jsonp.php/t/US_CategoryService.getList?page=1&num=60", "jsonp")
	req = robot.NewRequest("http://soft.chinabyte.com/416/13164916.shtml", "html", "", "GET", "", nil, nil, nil, nil)

	var dl robot.Downloader
	dl = NewHttpDownloader()

	var p *robot.Page
	p = dl.Download(req)

	//hp := p.GetHtmlParser()
	//fmt.Printf("%v", jsonMap)

	//fmt.Println(doc)
	p.GetBodyStr()
	body := p.GetBodyStr()
	fmt.Println(body)
}
