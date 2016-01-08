/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: 基于nsq的调度策略，可以实现分布式，爬虫有记忆功能
不建议使用nsq做调度，nsq是主动向蜘蛛推送url，而实际场景是蜘蛛主动向调度器获取url
*/
package scheduler

import (
	"encoding/json"

	"github.com/aosen/mlog"
	"github.com/aosen/robot"
	"github.com/nsqio/go-nsq"
)

type NsqScheduler struct {
	nsqaddr string
	topic   string
	channel string
	cfg     *nsq.Config
}

func NewNsqScheduler(nsqaddr, topic, channel string) *NsqScheduler {
	ns := &NsqScheduler{
		nsqaddr: nsqaddr,
		topic:   topic,
		channel: channel,
	}
	ns.Init()
	return ns
}

func (self *NsqScheduler) Init() {
	self.cfg = nsq.NewConfig()
}

func (self *NsqScheduler) Push(requ *robot.Request) {
	producer, err := nsq.NewProducer(self.nsqaddr, self.cfg)
	if err != nil {
		mlog.LogInst().LogError("NsqScheduler NewProducer Error: " + err.Error())
		return
	}
	defer producer.Stop()

	requJson, err := json.Marshal(requ)
	if err != nil {
		mlog.LogInst().LogError("NsqScheduler Json Error: " + err.Error())
		return
	}

	e := producer.Publish(self.topic, requJson)
	if e != nil {
		mlog.LogInst().LogError("NsqScheduler Push Error: " + e.Error())
		return
	}
}

func (self *NsqScheduler) Poll() *robot.Request {
	cs, err := nsq.NewConsumer(self.topic, self.channel, self.cfg)
	if err != nil {
		mlog.LogInst().LogError("NsqScheduler NewConsumer Error: " + err.Error())
	}
}
