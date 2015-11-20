package robot

/*
需要实现的爬虫接口
1. Downloader接口
*/

import ()

// The Downloader interface.
// You can implement the interface by implement function Download.
// Function Download need to return Page instance pointer that has request result downloaded from Request.
type Downloader interface {
	Download(req *Request) *Page
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
type ResourceManage interface {
	GetOne()
	FreeOne()
	Has() uint
	Left() uint
}
