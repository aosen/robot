/*
Author: Aosen
Data: 2016-01-19
QQ: 316052486
Desc: models
*/
package pipeline

import (
	"database/sql"
	"log"
	"time"

	"github.com/aosen/robot"
	"github.com/aosen/robot/example/www79xscom/utils"
	"github.com/astaxie/beedb"
)

//系统配置表
type System struct {
	Id int    `PK`
	K  string `orm:"size(50);unique"`
	V  string `orm:"size(100)"`
}

//一级分类表
type First struct {
	Id         int       `PK`
	Firstname  string    `orm:"size(20);unique"`
	Updatetime time.Time `orm:"type(date)"`
	Createtime time.Time `orm:"type(date)"`
}

//二级分类表
type Second struct {
	Id         int       `PK`
	Secondname string    `orm:"size(20);unique"`
	Updatetime time.Time `orm:"type(date)"`
	Createtime time.Time `orm:"type(date)"`
}

//小说简介表
type Novel struct {
	Id           int       `PK`
	Title        string    `orm:"size(200)"`
	Firstid      int       `orm:"index"`
	Secondid     int       `orm:"index"`
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
	Id            int `PK`
	Novelid       int
	Title         string    `orm:"size(200);index"`
	Firstid       int       `orm:"index"`
	Secondid      int       `orm:"index"`
	Chapter       int       `orm:"index"`
	Subtitle      string    `orm:"size(200);index"`
	Text          string    `orm:"type(text)"`
	Contentsource string    `orm:"size(200);index"`
	Createtime    time.Time `orm:"type(date)"`
}

//mysql pipeline
type PipelineMySQL struct {
	DB  *sql.DB
	ORM beedb.Model
	//图片存储路径
	ImageStore string
}

func NewPipelineMySQL(dbinfo string) *PipelineMySQL {
	db, _ := sql.Open("mysql", dbinfo)
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(30)
	if db.Ping() != nil {
		log.Fatal("连接数据库失败")
	}
	pm := new(PipelineMySQL)
	pm.DB = db
	pm.ORM = beedb.New(db)
	system := &System{}
	err := pm.ORM.Where("k=?", "imagestore").Find(system)
	if err != nil {
		log.Fatal("在mysql数据库system表中没有找到imagestore")
	}
	pm.ImageStore = system.V
	return pm
}

func (self *PipelineMySQL) Process(pageitems *robot.PageItems, task robot.Task) {
	//如果code＝“0” 则调用处理一级分类函数
	if code, ok := pageitems.GetItem("code"); ok {
		switch code {
		case "0":
			//处理一级分类
			self.firstProcess(pageitems, task)
		case "1":
			//处理二级分类
			self.secondProcess(pageitems, task)
		case "2":
			//下载图片
			self.imgProcess(pageitems, task)
		}
	}
}

//如果一级分类标签存在则略过，
//不存在则将一级分类标签插入数据库，并存储一级分类标签的id
func (self *PipelineMySQL) firstProcess(pageitems *robot.PageItems, task robot.Task) {
	if firstname, ok := pageitems.GetItem("first"); ok {
		first := &First{}
		err := self.ORM.Where("firstname=?", firstname).Find(first)
		//如果数据不存在 则创建
		if err != nil {
			first.Firstname = firstname
			first.Createtime = time.Now()
			first.Updatetime = time.Now()
			self.ORM.Save(first)
		}
	}
}

//如果二级分类存在则略过，不存在存储
func (self *PipelineMySQL) secondProcess(pageitems *robot.PageItems, task robot.Task) {
	if secondname, ok := pageitems.GetItem("second"); ok {
		second := &Second{}
		err := self.ORM.Where("secondname=?", secondname).Find(second)
		//如果数据不存在 则创建
		if err != nil {
			second.Secondname = secondname
			second.Createtime = time.Now()
			second.Updatetime = time.Now()
			self.ORM.Save(second)
		}
	}
}

//下载图片
func (self *PipelineMySQL) imgProcess(pageitems *robot.PageItems, task robot.Task) {
	if img, ok := pageitems.GetItem("img"); ok {
		filename, _ := utils.DownloadImage(img, self.ImageStore)
		pageitems.AddItem("picture", filename)
	}
}
