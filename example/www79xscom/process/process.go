/*
Author: Aosen
Data: 2016-01-19
QQ: 316052486
Desc: 页面处理类
*/

package process

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aosen/goutils"
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
		log.Println(p.Errormsg())
		return
	}

	meta := p.GetRequest().GetMeta()
	handler, ok := meta.(map[string]interface{})["handler"]
	//如果meta中没有handler处理方法，则说明是入口页面，否则直接执行对应callback
	if ok {
		switch handler {
		case "mainParse":
			self.mainParse(p)
		case "urlListParse":
			self.urlListParse(p)
		case "classParse":
			self.classParse(p)
		case "introParse":
			self.introParse(p)
		case "chaperParse":
			self.chaperParse(p)
		case "contentParse":
			self.contentParse(p)
		default:
			return
		}
	}
}

//主页解析
func (self *Www79xsComProcessor) mainParse(p *robot.Page) {
	//开始解析页面
	query := p.GetHtmlParser()
	query.Find(".subnav ul li a").Each(func(i int, s *goquery.Selection) {
		addr, _ := s.Attr("href")
		if addr == utils.GirlUrl {
			//p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+addr, map[string]string{"first": utils.GIRL}, self.urlListParse))
			p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+addr, map[string]string{"handler": "urlListParse", "first": utils.GIRL}))
		} else {
			p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+addr, map[string]string{"handler": "urlListParse", "first": utils.BOY}))
		}
	})
}

//获取分类页面的url list，并解析
func (self *Www79xsComProcessor) urlListParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]interface{})
	meta["handler"] = "classParse"
	//开始解析页面
	query := p.GetHtmlParser()
	//获取尾页addr
	lastaddr, ok := query.Find("tbody a").Last().Attr("href")
	if ok {
		//解析addr
		kv := goutils.GetKVInRelaPath(lastaddr)
		//url拼接
		//maxpage, _ := strconv.Atoi(kv["page"])
		maxpage := 0
		for i := 1; i <= maxpage; i++ {
			page := strconv.Itoa(i)
			p.AddTargetRequest(utils.InitRequest(
				"http://www.79xs.com/Book/ShowBookList.aspx?tclassid="+kv["tclassid"]+"&page="+page,
				meta))
		}
	} else {
		p.AddTargetRequest(utils.InitRequest(p.GetRequest().GetUrl(), meta))
	}
}

//分类列表解析
func (self *Www79xsComProcessor) classParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]interface{})
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
		tmp := make(map[string]interface{})
		tmp["first"] = meta["first"]
		tmp["second"] = second
		tmp["title"] = title
		tmp["author"] = author
		tmp["novelsource"] = novelsource
		tmp["handler"] = "introParse"
		p.AddTargetRequest(utils.InitRequest(novelsource, tmp))
	})
}

//解析小说详情页
func (self *Www79xsComProcessor) introParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]interface{})
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
	tmp["handler"] = "chaperParse"
	p.AddTargetRequest(utils.InitRequest(utils.BaseUrl+chaptersource, tmp))
}

//小说章节解析
func (self *Www79xsComProcessor) chaperParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]interface{})
	//开始解析页面
	query := p.GetHtmlParser()
	query.Find(".insert_list li").Each(func(i int, s *goquery.Selection) {
		tmp := utils.MapCopy(meta)
		tmp["chapter"] = strconv.Itoa(i)
		tmp["subtitle"] = s.Find("strong a").Text()
		addr, _ := s.Find("strong a").Attr("href")
		tmp["contenturl"] = p.GetRequest().GetBaseUrl() + addr
		tmp["handler"] = "contentParse"
		//检测contenturl, 如果数据库中存在，则跳过本次抓取，如果不存在则将url加入调度队列
		//这个需求有时间再做
		if len(tmp["subtitle"].(string)) != 0 {
			p.AddTargetRequest(utils.InitRequest(tmp["contenturl"].(string), tmp))
		}
	})
}

//小说内容解析
func (self *Www79xsComProcessor) contentParse(p *robot.Page) {
	meta := p.GetRequest().GetMeta().(map[string]interface{})
	//开始解析页面
	query := p.GetHtmlParser()
	html, _ := query.Find(".contentbox").Html()
	meta["content"] = strings.Replace(strings.Replace(html, "<br/><br/>", "\n", -1), "<br/>", "\n", -1)
	p.AddField("code", "0")
	for k, v := range meta {
		p.AddField(k, v.(string))
	}
}

func (self *Www79xsComProcessor) Finish() {
}
