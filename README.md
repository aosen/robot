# robot
Golang爬虫框架。

##robot分四个接口：
* 下载模块(Downloader)
* 页面处理模块(PageProcesser)
* 任务调度模块(Scheduler)
* pipline

其中下载模块，调度模块已经实现，需要开发者自行实现页面处理规则与pipline

##开发进度
* 2015-01-14 增加爬虫池支持，提高爬虫系统性能 ***Done***
* 2015-01-07 优化目录结构 ***Done***
