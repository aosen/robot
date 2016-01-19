/*
Author: Aosen
Data: 2016-01-19
QQ: 316052486
Desc: 页面处理类
*/

package process

import (
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
