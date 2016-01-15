/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: 基于redis的调度策略,可以实现分布式 爬虫有记忆功能
*/

package scheduler

import (
	"encoding/json"

	"github.com/aosen/mlog"
	"github.com/aosen/robot"
	"github.com/garyburd/redigo/redis"
)

type RedisSchedulerOptions struct {
	//请求对流的名称
	RequestList string
	//已经抓取的url列表
	UrlList string
	//redis连接地址
	RedisAddr string
	//最大连接数
	MaxConn int
	MaxIdle int
	//是否做url去重
	ForbiddenDuplicateUrl bool
}

type RedisScheduler struct {
	requestList           string
	urlList               string
	redisAddr             string
	redisPool             *redis.Pool
	maxConn               int
	maxIdle               int
	forbiddenDuplicateUrl bool
}

func NewRedisScheduler(options RedisSchedulerOptions) *RedisScheduler {
	rs := &RedisScheduler{
		redisAddr:             options.RedisAddr,
		forbiddenDuplicateUrl: options.ForbiddenDuplicateUrl,
		maxConn:               options.MaxConn,
		maxIdle:               options.MaxIdle,
		requestList:           options.RequestList,
		urlList:               options.UrlList,
	}
	rs = rs.Init()
	return rs
}

func (self *RedisScheduler) Init() *RedisScheduler {
	self.redisPool = redis.NewPool(self.newConn, self.maxIdle)
	self.redisPool.MaxActive = self.maxConn
	return self
}

func (self *RedisScheduler) newConn() (c redis.Conn, err error) {
	c, err = redis.Dial("tcp", self.redisAddr)
	if err != nil {
		panic(err)
	}
	return
}

func (self *RedisScheduler) Push(requ *robot.Request) {
	requJson, err := json.Marshal(requ)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
		return
	}

	conn := self.redisPool.Get()
	defer conn.Close()

	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
		return
	}
	if self.forbiddenDuplicateUrl {
		urlExist, err := conn.Do("HGET", self.urlList, requ.GetUrl())
		if err != nil {
			mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
			return
		}
		if urlExist != nil {
			return
		}
		conn.Do("MULTI")
		_, err = conn.Do("HSET", self.urlList, requ.GetUrl(), 1)
		if err != nil {
			mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
			conn.Do("DISCARD")
			return
		}
	}
	_, err = conn.Do("RPUSH", self.requestList, requJson)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Push Error: " + err.Error())
		if self.forbiddenDuplicateUrl {
			conn.Do("DISCARD")
		}
		return
	}

	if self.forbiddenDuplicateUrl {
		conn.Do("EXEC")
	}
}

func (self *RedisScheduler) Poll() *robot.Request {
	conn := self.redisPool.Get()
	defer conn.Close()

	length, err := self.count()
	if err != nil {
		return nil
	}
	if length <= 0 {
		mlog.LogInst().LogError("RedisScheduler Poll length 0")
		return nil
	}
	buf, err := conn.Do("LPOP", self.requestList)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Poll Error: " + err.Error())
		return nil
	}

	requ := &robot.Request{}

	err = json.Unmarshal(buf.([]byte), requ)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Poll Error: " + err.Error())
		return nil
	}

	return requ
}

func (self *RedisScheduler) Count() int {
	var length int
	var err error

	length, err = self.count()
	if err != nil {
		return 0
	}

	return length
}

func (self *RedisScheduler) count() (int, error) {
	conn := self.redisPool.Get()
	defer conn.Close()
	length, err := conn.Do("LLEN", self.requestList)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Count Error: " + err.Error())
		return 0, err
	}
	return int(length.(int64)), nil
}
