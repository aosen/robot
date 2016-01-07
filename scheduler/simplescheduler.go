/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc:最简单的调度策略,当request超过1024会发生阻塞.
*/
package scheduler

import "github.com/aosen/robot"

type SimpleScheduler struct {
	queue chan *robot.Request
}

func NewSimpleScheduler() *SimpleScheduler {
	ch := make(chan *robot.Request, 1024)
	return &SimpleScheduler{ch}
}

func (this *SimpleScheduler) Push(requ *robot.Request) {
	this.queue <- requ
}

func (this *SimpleScheduler) Poll() *robot.Request {
	if len(this.queue) == 0 {
		return nil
	} else {
		return <-this.queue
	}
}

func (this *SimpleScheduler) Count() int {
	return len(this.queue)
}
