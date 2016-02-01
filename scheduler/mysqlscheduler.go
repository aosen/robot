/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: 基于mysql的调度策略,可以实现分布式 爬虫有记忆功能
*/

package scheduler

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/aosen/robot"
	"github.com/astaxie/beego/orm"
)

type Requestlist struct {
	Id int64
	//爬虫名称
	Name string `orm:"size(20)"`
	//请求
	Requ       string    `orm:"type(text)"`
	Createtime time.Time `orm:"type(date)"`
}

type MysqlScheduler struct {
	//爬虫名称
	name   string
	dbinfo string
	locker *sync.Mutex
}

func NewMysqlScheduler(name, dbinfo string) *MysqlScheduler {
	my := &MysqlScheduler{
		name:   name,
		locker: new(sync.Mutex),
	}
	my = my.Init(dbinfo)
	return my
}

func (self *MysqlScheduler) Init(dbinfo string) *MysqlScheduler {
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDataBase("scheduler", "mysql", dbinfo)
	orm.RegisterModel(new(Requestlist))
	return self
}

func (self *MysqlScheduler) Push(requ *robot.Request) {
	self.locker.Lock()
	defer self.locker.Unlock()
	requJson, err := json.Marshal(requ)
	if err != nil {
		log.Println("Scheduler Push Error: " + err.Error())
		return
	}
	req := &Requestlist{
		Name:       self.name,
		Requ:       string(requJson),
		Createtime: time.Now(),
	}
	o := orm.NewOrm()
	o.Using("scheduler")
	_, err = o.Insert(req)
	if err != nil {
		log.Println("Push Error:" + err.Error())
	}
}

func (self *MysqlScheduler) Poll() *robot.Request {
	self.locker.Lock()
	defer self.locker.Unlock()
	o := orm.NewOrm()
	o.Using("scheduler")
	req := &Requestlist{}
	err := o.QueryTable("requestlist").Filter("name", self.name).Limit(1).One(req)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		log.Printf("Returned Multi Rows Not One")
		return nil
	}
	if err == orm.ErrNoRows {
		// 没有找到记录
		//log.Printf("Not row found")
		return nil
	}
	//删除数据
	if _, err := o.Raw("DELETE FROM requestlist WHERE id = ?", req.Id).Exec(); err != nil {
		//if _, err := o.Delete(&Requestlist{Id: req.Id}); err != nil {
		log.Println("Delete data error:" + err.Error())
	}

	r := &robot.Request{}
	err = json.Unmarshal([]byte(req.Requ), r)
	if err != nil {
		log.Println("Poll error:" + err.Error())
		return nil
	}
	return r
}

func (self *MysqlScheduler) Count() int {
	self.locker.Lock()
	defer self.locker.Unlock()
	o := orm.NewOrm()
	o.Using("scheduler")
	cnt, err := o.QueryTable("requestlist").Filter("name", self.name).Count()
	if err != nil {
		log.Println("count err: " + err.Error())
	}
	return int(cnt)
}
