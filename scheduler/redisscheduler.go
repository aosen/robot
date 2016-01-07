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

type RedisScheduler struct {
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
	return this
}

func (this *RedisScheduler) newConn() (c redis.Conn, err error) {
	c, err = redis.Dial("tcp", this.redisAddr)
	if err != nil {
		panic(err)
	}
	return
}

func (this *RedisScheduler) Push(requ *robot.Request) {
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

func (this *RedisScheduler) Poll() *robot.Request {
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

	requ := &robot.Request{}

	err = json.Unmarshal(buf.([]byte), requ)
	if err != nil {
		mlog.LogInst().LogError("RedisScheduler Poll Error: " + err.Error())
		return nil
	}

	return requ
}

func (this *RedisScheduler) Count() int {
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
