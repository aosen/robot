/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc:
基于队列的调度策略队列的容量可动态增加，不会产生阻塞，但无法实现分布式.
*/
package scheduler

import (
	"container/list"
	"crypto/md5"
	"sync"

	"github.com/aosen/robot"
)

type QueueScheduler struct {
	locker *sync.Mutex
	rm     bool
	rmKey  map[[md5.Size]byte]*list.Element
	queue  *list.List
}

func NewQueueScheduler(rmDuplicate bool) *QueueScheduler {
	queue := list.New()
	rmKey := make(map[[md5.Size]byte]*list.Element)
	locker := new(sync.Mutex)
	return &QueueScheduler{rm: rmDuplicate, queue: queue, rmKey: rmKey, locker: locker}
}

func (self *QueueScheduler) Push(requ *robot.Request) {
	self.locker.Lock()
	var key [md5.Size]byte
	if self.rm {
		key = md5.Sum([]byte(requ.GetUrl()))
		if _, ok := self.rmKey[key]; ok {
			self.locker.Unlock()
			return
		}
	}
	e := self.queue.PushBack(requ)
	if self.rm {
		self.rmKey[key] = e
	}
	self.locker.Unlock()
}

func (self *QueueScheduler) Poll() *robot.Request {
	self.locker.Lock()
	if self.queue.Len() <= 0 {
		self.locker.Unlock()
		return nil
	}
	e := self.queue.Front()
	requ := e.Value.(*robot.Request)
	key := md5.Sum([]byte(requ.GetUrl()))
	self.queue.Remove(e)
	if self.rm {
		delete(self.rmKey, key)
	}
	self.locker.Unlock()
	return requ
}

func (self *QueueScheduler) Count() int {
	self.locker.Lock()
	len := self.queue.Len()
	self.locker.Unlock()
	return len
}
