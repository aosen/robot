/*
Author: Aosen
Data: 2016-01-19
QQ: 316052486
Desc: 页面处理类
*/

package process

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/mlog"
	"github.com/aosen/robot"
	"github.com/aosen/robot/example/www79xscom/utils"
)

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
		if addr == utils.GirlUrl {
			p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+addr, map[string]string{"first": utils.GIRL}, self.urlListParse))
		} else {
			p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+addr, map[string]string{"first": utils.BOY}, self.urlListParse))
		}
	})
}

//获取分类页面的url list，并解析
func (self *Www79xsComProcessor) urlListParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta()
	//开始解析页面
	query := p.GetHtmlParser()
	//获取尾页addr
	//lastaddr, ok := query.Find("tbody a").Last().Attr("href")
	_, ok := query.Find("tbody a").Last().Attr("href")
	if ok {
		//解析addr
		/* 测试完成后将注释取消
		kv := goutils.GetKVInRelaPath(lastaddr)
		//url拼接
		maxpage, _ := strconv.Atoi(kv["page"])
		for i := 1; i <= maxpage; i++ {
			page := strconv.Itoa(i)
			p.AddTargetRequest(utils.InitRequest(
				"http://www.79xs.com/Book/ShowBookList.aspx?tclassid="+kv["tclassid"]+"&page="+page,
				meta.(map[string]string),
				self.classParse))
		}
		*/
	} else {
		p.AddTargetRequest(utils.InitRequest(p.GetRequest().GetUrl(), meta.(map[string]string), self.classParse))
	}
}

//分类列表解析
func (self *Www79xsComProcessor) classParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]string)
	//开始解析页面
	query := p.GetHtmlParser()
	query.Find("div .yl_nr_lt2 ul").Each(func(i int, s *goquery.Selection) {
		//获取二级分类, 小说标题，作者
		second := s.Find(".ynl2 a").Text()
		title := s.Find(".ynl3 a").Eq(1).Text()
		author := s.Find(".ynl6 a").Text()
		novelsource := utils.BaseUrl + func() string {
			addr, _ := s.Find(".ynl3 a").Eq(1).Attr("href")
			return addr
		}()
		tmp := make(map[string]string)
		tmp["first"] = meta["first"]
		tmp["second"] = second
		tmp["title"] = title
		tmp["author"] = author
		tmp["novelsource"] = novelsource
		p.AddTargetRequest(utils.InitRequest(novelsource, tmp, self.introParse))
	})
}

//解析小说详情页
func (self *Www79xsComProcessor) introParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]string)
	//开始解析页面
	query := p.GetHtmlParser()
	intro := query.Find("#info h3 p").Eq(1).Text()
	img, _ := query.Find(".img img").Attr("src")
	// 小说章节列表地址
	chaptersource, _ := query.Find(".b1 a").Attr("href")
	tmp := utils.MapCopy(meta)
	tmp["introduction"] = intro
	tmp["img"] = utils.BaseUrl + img
	tmp["chaptersource"] = utils.BaseUrl + chaptersource
	p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+chaptersource, tmp, self.chaperParse))
}

//小说章节解析
func (self *Www79xsComProcessor) chaperParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]string)
	//开始解析页面
	query := p.GetHtmlParser()
	query.Find(".insert_list li").Each(func(i int, s *goquery.Selection) {
		tmp := utils.MapCopy(meta)
		tmp["chapter"] = strconv.Itoa(i)
		tmp["subtitle"] = s.Find("strong a").Text()
		addr, _ := s.Find("strong a").Attr("href")
		tmp["contenturl"] = p.GetRequest().GetBaseUrl() + addr
		//检测contenturl, 如果数据库中存在，则跳过本次抓取，如果不存在则将url加入调度队列
		//这个需求有时间再做
		p.AddTargetRequest(utils.InitRequest(tmp["contenturl"], tmp, self.contentParse))
	})
}

//小说内容解析
func (self *Www79xsComProcessor) contentParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]string)
	//开始解析页面
	query := p.GetHtmlParser()
	html, _ := query.Find(".contentbox").Html()
	meta["content"] = strings.Replace(strings.Replace(html, "<br/><br/>", "\n", -1), "<br/>", "\n", -1)
	p.AddField("code", "0")
	for k, v := range meta {
		p.AddField(k, v)
	}

}

func (self *Www79xsComProcessor) Finish() {
}
