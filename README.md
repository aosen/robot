# robot
Golang爬虫框架。即可做通用爬虫也可做垂直爬虫

##使用过程中遇到任何问题，请联系QQ: 316052486

##思维脑图

![思维脑图](http://i8.tietuku.com/b04bfbc8c05faf28.png "思维脑图")

##robot分四个接口：
* 下载模块(Downloader)
* 页面处理模块(PageProcesser)
* 任务调度模块(Scheduler)
* pipline

其中下载模块，调度模块已经实现，需要开发者自行实现页面处理规则与pipline

##robot爬虫框架优势
* 完全接口话，你可以随意定制下载器，页面处理器，调度器，pipeline
* 爬虫池思想，稳定高效

##使用说明
###初始化爬虫
```Golang

//爬虫选项设置
options := robot.SpiderOptions{
    TaskName:      spidername,
    PageProcesser: process.NewWww79xsComProcessor(),
    Downloader:    downloader.NewHttpDownloader("text/html; charset=gb2312"),
    Scheduler:     scheduler.NewMysqlScheduler(spidername, dbinfo),
    //Scheduler: scheduler.NewQueueScheduler(false),
    Pipelines: []robot.Pipeline{pipeline.NewPipelineMySQL(dbinfo)},
    //设置资源管理器，资源池容量为100
    ResourceManage: resource.NewSpidersPool(100, nil),
}

//爬虫初始化
sp := robot.NewSpider(options)
//增加根url
sp.AddRequest(utils.InitRequest(start_url, map[string]string{
    "handler": "mainParse",
}))
go sp.Run()
```
###详细用列见
example下

##开发进度
* 2015-01-18 增加爬虫案例： example/www79xscom/ ***Done***
* 2015-01-18 Request结构体中增加回调函数支持，更好支持垂直爬虫的实现,具体实例见 example/www79xscom/  ***Done***
* 2015-01-14 增加爬虫池支持，提高爬虫系统性能 ***Done***
* 2015-01-07 优化目录结构 ***Done***
