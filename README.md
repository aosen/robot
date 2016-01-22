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

##开发进度
* 2015-01-18 增加爬虫案例： example/www79xscom/ ***Done***
* 2015-01-18 Request结构体中增加回调函数支持，更好支持垂直爬虫的实现,具体实例见 example/www79xscom/  ***Done***
* 2015-01-14 增加爬虫池支持，提高爬虫系统性能 ***Done***
* 2015-01-07 优化目录结构 ***Done***
