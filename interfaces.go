package robot

/*
需要实现的爬虫接口
1. Downloader接口
*/

// The Downloader interface.
// You can implement the interface by implement function Download.
// Function Download need to return Page instance pointer that has request result downloaded from Request.
type Downloader interface {
	Download(req *Request) *Page
}

//url调度接口
type Scheduler interface {
	Push(requ *Request)
	Poll() *Request
	Count() int
}

// The Task represents interface that contains environment variables.
// It inherits by Spider.
type Task interface {
	Taskname() string
}

// 页面下载后的处理接口，需要开发者自己实现
type PageProcesser interface {
	Process(p *Page)
	Finish()
}

// The interface Pipeline can be implemented to customize ways of persistent.
// 最终抓取数据流向，需开发者自己实现，pipeline文件夹下有例子
type Pipeline interface {
	// The Process implements result persistent.
	// The items has the result be crawled.
	// The t has informations of this crawl task.
	Process(items *PageItems, t Task)
}

// The interface CollectPipeline recommend result in process's memory temporarily.
type CollectPipeline interface {
	Pipeline

	// The GetCollected returns result saved in in process's memory temporarily.
	GetCollected() []*PageItems
}

//资源管理接口
// ResourceManage is an interface that who want implement an management object can realize these functions.
/*
type ResourceManage interface {
	GetOne()
	FreeOne()
	Has() uint
	Left() uint
}
*/
type ResourceManage interface {
	//启动资源管理器
	Start()
	//释放资源管理器
	Free()
	//向资源管理器中添加任务
	AddTask(func(*Request), *Request)
	//获取资源管理器中的资源量
	Has() int
}
