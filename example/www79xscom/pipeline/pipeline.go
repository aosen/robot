/*
Author: Aosen
Data: 2016-01-19
QQ: 316052486
Desc: models
*/
package pipeline

import (
	"log"
	"strconv"
	"time"

	"github.com/aosen/robot"
	"github.com/aosen/robot/example/www79xscom/utils"
	"github.com/astaxie/beego/orm"
)

//系统配置表
type System struct {
	Id int64
	K  string `orm:"size(50);unique"`
	V  string `orm:"size(100)"`
}

//一级分类表
type First struct {
	Id         int64
	Firstname  string    `orm:"size(20);unique"`
	Updatetime time.Time `orm:"type(date)"`
	Createtime time.Time `orm:"type(date)"`
}

//二级分类表
type Second struct {
	Id         int64
	Secondname string    `orm:"size(20);unique"`
	Updatetime time.Time `orm:"type(date)"`
	Createtime time.Time `orm:"type(date)"`
}

//小说简介表
type Novel struct {
	Id           int64
	Title        string    `orm:"size(200)"`
	Firstid      int64     `orm:"index"`
	Secondid     int64     `orm:"index"`
	Author       string    `orm:"size(50);index"`
	Introduction string    `orm:"type(text)"`
	Picture      string    `orm:"size(200)"`
	Novelsource  string    `orm:"size(200);unique"`
	Novelpv      int       `orm:"default(0)"`
	Novelcollect int       `orm:"default(0)"`
	Createtime   time.Time `orm:"type(date)"`
}

//小说内容表
type Content struct {
	Id            int64
	Novelid       int64
	Title         string    `orm:"size(200);index"`
	Firstid       int64     `orm:"index"`
	Secondid      int64     `orm:"index"`
	Chapter       int       `orm:"index"`
	Subtitle      string    `orm:"size(200);index"`
	Text          string    `orm:"type(text)"`
	Contentsource string    `orm:"size(200);index"`
	Createtime    time.Time `orm:"type(date)"`
}

//mysql pipeline
type PipelineMySQL struct {
	ImageStore string
}

func NewPipelineMySQL(dbinfo string) *PipelineMySQL {
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDataBase("default", "mysql", dbinfo)
	orm.RegisterModel(new(First), new(Second), new(System), new(Novel), new(Content))
	o := orm.NewOrm()
	system := &System{}
	system.K = "imagestore"
	err := o.Read(system, "K")
	if err != nil {
		log.Fatal("在mysql数据库system表中没有找到imagestore")
	}
	return &PipelineMySQL{
		ImageStore: system.V,
	}
}

func (self *PipelineMySQL) Process(pageitems *robot.PageItems, task robot.Task) {
	if code, ok := pageitems.GetItem("code"); ok && code == "0" {
		firstid := self.firstProcess(pageitems, task)
		secondid := self.secondProcess(pageitems, task)
		picname := self.imgProcess(pageitems, task)
		if novelid, err := self.novelProcess(pageitems, task, firstid, secondid, picname); err == nil {
			self.contentProcess(pageitems, task, novelid, firstid, secondid)
			log.Println(firstid, secondid, picname, novelid)
		} else {
			log.Println(err.Error())
		}
	}
}

//如果一级分类标签存在则略过，
//不存在则将一级分类标签插入数据库，并存储一级分类标签的id
func (self *PipelineMySQL) firstProcess(pageitems *robot.PageItems, task robot.Task) int64 {
	if firstname, ok := pageitems.GetItem("first"); ok {
		o := orm.NewOrm()
		first := &First{
			Firstname:  firstname,
			Createtime: time.Now(),
			Updatetime: time.Now(),
		}
		if _, id, err := o.ReadOrCreate(first, "firstname"); err == nil {
			return id
		}

	}
	return -1
}

//如果二级分类存在则略过，不存在存储
func (self *PipelineMySQL) secondProcess(pageitems *robot.PageItems, task robot.Task) int64 {
	if secondname, ok := pageitems.GetItem("second"); ok {
		o := orm.NewOrm()
		second := &Second{
			Secondname: secondname,
			Createtime: time.Now(),
			Updatetime: time.Now(),
		}
		//如果数据不存在 则创建
		if _, id, err := o.ReadOrCreate(second, "secondname"); err == nil {
			return id
		}
	}
	return -1
}

//下载图片
func (self *PipelineMySQL) imgProcess(pageitems *robot.PageItems, task robot.Task) string {
	if img, ok := pageitems.GetItem("img"); ok {
		filename, _ := utils.DownloadImage(img, self.ImageStore)
		return filename
	}
	return ""
}

//添加小说表
func (self *PipelineMySQL) novelProcess(pageitems *robot.PageItems, task robot.Task, firstid, secondid int64, picname string) (int64, error) {
	items := pageitems.GetAll()
	o := orm.NewOrm()
	novel := &Novel{
		Title:        items["title"],
		Firstid:      firstid,
		Secondid:     secondid,
		Author:       items["author"],
		Introduction: items["introduction"],
		Picture:      picname,
		Novelsource:  items["novelsource"],
		Createtime:   time.Now(),
	}
	//如果数据不存在 则创建
	_, id, err := o.ReadOrCreate(novel, "novelsource")
	return id, err
}

func (self *PipelineMySQL) contentProcess(pageitems *robot.PageItems, task robot.Task, novelid, firstid, secondid int64) {
	items := pageitems.GetAll()
	chapter, _ := strconv.Atoi(items["chapter"])
	o := orm.NewOrm()
	content := &Content{
		Novelid:       novelid,
		Title:         items["title"],
		Firstid:       firstid,
		Secondid:      secondid,
		Chapter:       chapter,
		Subtitle:      items["subtitle"],
		Text:          items["content"],
		Contentsource: items["contenturl"],
		Createtime:    time.Now(),
	}
	//如果数据不存在 则创建
	o.ReadOrCreate(content, "contentsource")
}
