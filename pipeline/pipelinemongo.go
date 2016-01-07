/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: mongodb实现的pipeline 仅供参考，详细pipeline的实现还需开发者自行实现
*/

package pipeline

import (
	"github.com/aosen/robot"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//mongo pipline 的例子，仅供参考，需要开发者自己实现
type PipelineMongo struct {
	session           *mgo.Session
	url               string
	mongoDBName       string //数据库名
	mongoDBCollection string //集合名
	c                 *mgo.Collection
	items             interface{} //存储的items结构体类型, 用于后期的反射
}

type User struct {
	Id_  bson.ObjectId `bson:"_id"`
	Name string        `bson:"name"`
	Age  int           `bson:"age"`
}

func NewPipelineMongo(url, db, collection string) *PipelineMongo {
	session, err := mgo.Dial(url)
	if err != nil {
		panic("open mongodb file:" + err.Error())
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return &PipelineMongo{
		session:           session,
		url:               url,
		mongoDBName:       db,
		mongoDBCollection: collection,
		c:                 session.DB(db).C(collection),
	}
}

func (self *PipelineMongo) Process(pageitems *robot.PageItems, task robot.Task) {
	err := self.c.Insert(&User{
		Id_:  bson.NewObjectId(),
		Name: "Jimmy Kuu",
		Age:  33,
	})
	if err != nil {
		panic(err)
	}
}
