package robot

import (
	"container/list"
	"crypto/md5"
	"encoding/json"
	"github.com/aosen/mlog"
	"github.com/garyburd/redigo/redis"
	"sync"
)

type Scheduler interface {
	Push(requ *Request)
	Poll() *Request
	Count() int
}

//最简单的调度策略
//当request超过1024会发生阻塞
type SimpleScheduler struct {
	queue chan *Request
}

func NewSimpleScheduler() *SimpleScheduler {
	ch := make(chan *Request, 1024)
	return &SimpleScheduler{ch}
}

func (this *SimpleScheduler) Push(requ *Request) {
	this.queue <- requ
}

func (this *SimpleScheduler) Poll() *Request {
	if len(this.queue) == 0 {
		return nil
	} else {
		return <-this.queue
	}
}

func (this *SimpleScheduler) Count() int {
	return len(this.queue)
}

//基于队列的调度策略
//队列的容量可动态增加，不会产生阻塞，但无法实现分布式
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

func (this *QueueScheduler) Push(requ *Request) {
	this.locker.Lock()
	var key [md5.Size]byte
	if this.rm {
		key = md5.Sum([]byte(requ.GetUrl()))
		if _, ok := this.rmKey[key]; ok {
			this.locker.Unlock()
			return
		}
	}
	e := this.queue.PushBack(requ)
	if this.rm {
		this.rmKey[key] = e
	}
	this.locker.Unlock()
}

func (this *QueueScheduler) Poll() *Request {
	this.locker.Lock()
	if this.queue.Len() <= 0 {
		this.locker.Unlock()
		return nil
	}
	e := this.queue.Front()
	requ := e.Value.(*Request)
	key := md5.Sum([]byte(requ.GetUrl()))
	this.queue.Remove(e)
	if this.rm {
		delete(this.rmKey, key)
	}
	this.locker.Unlock()
	return requ
}

func (this *QueueScheduler) Count() int {
	this.locker.Lock()
	len := this.queue.Len()
	this.locker.Unlock()
	return len
}

//基于redis的调度策略
//可以实现分布式 爬虫有记忆功能
type RedisScheduler struct {
	locker                *sync.Mutex
	requestList           string
	urlList               string
	redisAddr             string
	redisPool             *redis.Pool
	maxConn               int
	maxIdle               int
	forbiddenDuplicateUrl bool
	queueMax              int
}

func NewRedisScheduler(addr string, maxConn, maxIdle int, forbiddenDuplicateUrl bool) *RedisScheduler {
	rs := &RedisScheduler{
		redisAddr:             addr,
		forbiddenDuplicateUrl: forbiddenDuplicateUrl,
		maxConn:               maxConn,
		maxIdle:               maxIdle,
		requestList:           "robot_request",
		urlList:               "robot_url",
	}
	rs = rs.Init()
	return rs
}

func (this *RedisScheduler) Init() *RedisScheduler {
	this.redisPool = redis.NewPool(this.newConn, this.maxIdle)
	this.redisPool.MaxActive = this.maxConn
	this.locker = new(sync.Mutex)
	return this
}

func (this *RedisScheduler) newConn() (c redis.Conn, err error) {
	c, err = redis.Dial("tcp", this.redisAddr)
	if err != nil {
		panic(err)
	}
	return
}
func (this *RedisScheduler) Push(requ *Request) {
	this.locker.Lock()
	defer this.locker.Unlock()

	requJson, err := json.Marshal(requ)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
		return
	}

	conn := this.redisPool.Get()
	defer conn.Close()

	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
		return
	}
	if this.forbiddenDuplicateUrl {
		urlExist, err := conn.Do("HGET", this.urlList, requ.GetUrl())
		if err != nil {
			mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
			return
		}
		if urlExist != nil {
			return
		}

		conn.Do("MULTI")
		_, err = conn.Do("HSET", this.urlList, requ.GetUrl(), 1)
		if err != nil {
			mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
			conn.Do("DISCARD")
			return
		}
	}
	_, err = conn.Do("RPUSH", this.requestList, requJson)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
		if this.forbiddenDuplicateUrl {
			conn.Do("DISCARD")
		}
		return
	}

	if this.forbiddenDuplicateUrl {
		conn.Do("EXEC")
	}
}

func (this *RedisScheduler) Poll() *Request {
	this.locker.Lock()
	defer this.locker.Unlock()

	conn := this.redisPool.Get()
	defer conn.Close()

	length, err := this.count()
	if err != nil {
		return nil
	}
	if length <= 0 {
		mlog.LogInst().LogError("RedisScheduler Poll length 0")
		return nil
	}
	buf, err := conn.Do("LPOP", this.requestList)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Poll Error: " + err.Error())
		return nil
	}

	requ := &Request{}

	err = json.Unmarshal(buf.([]byte), requ)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Poll Error: " + err.Error())
		return nil
	}

	return requ
}

func (this *RedisScheduler) Count() int {
	this.locker.Lock()
	defer this.locker.Unlock()
	var length int
	var err error

	length, err = this.count()
	if err != nil {
		return 0
	}

	return length
}

func (this *RedisScheduler) count() (int, error) {
	conn := this.redisPool.Get()
	defer conn.Close()
	length, err := conn.Do("LLEN", this.requestList)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Count Error: " + err.Error())
		return 0, err
	}
	return int(length.(int64)), nil
}
