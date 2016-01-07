/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: 仅供参考，详细pipeline的实现还需开发者自行实现
*/

package pipeline

import "github.com/aosen/robot"

type PipelineConsole struct {
}

func NewPipelineConsole() *PipelineConsole {
	return &PipelineConsole{}
}

func (self *PipelineConsole) Process(items *robot.PageItems, t robot.Task) {
	println("----------------------------------------------------------------------------------------------")
	println("Crawled url :\t" + items.GetRequest().GetUrl() + "\n")
	println("Crawled result : ")
	for key, value := range items.GetAll() {
		println(key + "\t:\t" + value)
	}
}
