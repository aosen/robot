package robot

/*
pipline接口的实现，一般生产环境需要根据持久化需求重新实现pipline接口
*/

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
)

type CollectPipelinePageItems struct {
	collector []*PageItems
}

func NewCollectPipelinePageItems() *CollectPipelinePageItems {
	collector := make([]*PageItems, 0)
	return &CollectPipelinePageItems{collector: collector}
}

func (self *CollectPipelinePageItems) Process(items *PageItems, t Task) {
	self.collector = append(self.collector, items)
}

func (self *CollectPipelinePageItems) GetCollected() []*PageItems {
	return self.collector
}

type PipelineConsole struct {
}

func NewPipelineConsole() *PipelineConsole {
	return &PipelineConsole{}
}

func (self *PipelineConsole) Process(items *PageItems, t Task) {
	println("----------------------------------------------------------------------------------------------")
	println("Crawled url :\t" + items.GetRequest().GetUrl() + "\n")
	println("Crawled result : ")
	for key, value := range items.GetAll() {
		println(key + "\t:\t" + value)
	}
}

type PipelineFile struct {
	file *os.File

	path string
}

func NewPipelineFile(path string) *PipelineFile {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic("File '" + path + "' in PipelineFile open failed.")
	}
	return &PipelineFile{path: path, file: file}
}

func (self *PipelineFile) Process(items *PageItems, t Task) {
	self.file.WriteString("----------------------------------------------------------------------------------------------\n")
	self.file.WriteString("Crawled url :\t" + items.GetRequest().GetUrl() + "\n")
	self.file.WriteString("Crawled result : \n")
	for key, value := range items.GetAll() {
		self.file.WriteString(key + "\t:\t" + value + "\n")
	}
}

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

func (self *PipelineMongo) Process(pageitems *PageItems, task Task) {
	err := self.c.Insert(&User{
		Id_:  bson.NewObjectId(),
		Name: "Jimmy Kuu",
		Age:  33,
	})
	if err != nil {
		panic(err)
	}
}
