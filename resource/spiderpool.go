/*
Author: Aosen
Data: 2016-01-14
QQ: 316052486
Desc:
管理爬虫资源, 一个爬虫对应一个协程
*/
package resource

import "github.com/aosen/robot"

//任务类型
type task struct {
	//方法
	method func(*robot.Request)
	//参数
	req *robot.Request
}

func CreateTask(method func(*robot.Request), req *robot.Request) task {
	return task{
		method: method,
		req:    req,
	}
}

type SpidersPool struct {
	//任务队列
	queue chan task
	//协程池容量 和队列的容量
	//也就是一个任务对应一个蜘蛛，当蜘蛛处理慢的时候，队列只好阻塞
	total int
	//完成后的回调函数
	finishCallback func()
	//爬虫池是否被关闭的标示
	closed chan bool
}

//新建爬虫池
func NewSpidersPool(total int, callback func()) *SpidersPool {
	sp := new(SpidersPool)
	sp.total = total
	sp.finishCallback = callback
	sp.queue = make(chan task, total)
	sp.closed = make(chan bool)
	return sp
}

// 爬虫池开始接任务
func (self *SpidersPool) Start() {
	// 开启total个goroutine
	for i := 0; i < self.total; i++ {
		go func() {
			for {
				task, ok := <-self.queue
				if !ok {
					break
				}
				task.method(task.req)
			}
		}()
	}

	//知道爬虫池被关闭才继续向下运行
	<-self.closed
	close(self.closed)

	// 所有任务都执行完成，回调函数
	if self.finishCallback != nil {
		self.finishCallback()
	}
}

// 释放爬虫池
func (self *SpidersPool) Free() {
	close(self.queue)
	self.closed <- true
}

// 添加任务
func (self *SpidersPool) AddTask(method func(req *robot.Request), req *robot.Request) {
	self.queue <- CreateTask(method, req)
}

// 获取当前任务数
func (self *SpidersPool) Has() int {
	return len(self.queue)
}
