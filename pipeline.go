package robot

/*
pipline接口的实现，一般生产环境需要根据持久化需求重新实现pipline接口
*/

import (
	"os"
)

type CollectPipelinePageItems struct {
	collector []*PageItems
}

func NewCollectPipelinePageItems() *CollectPipelinePageItems {
	collector := make([]*PageItems, 0)
	return &CollectPipelinePageItems{collector: collector}
}

func (this *CollectPipelinePageItems) Process(items *PageItems, t Task) {
	this.collector = append(this.collector, items)
}

func (this *CollectPipelinePageItems) GetCollected() []*PageItems {
	return this.collector
}

type PipelineConsole struct {
}

func NewPipelineConsole() *PipelineConsole {
	return &PipelineConsole{}
}

func (this *PipelineConsole) Process(items *PageItems, t Task) {
	println("----------------------------------------------------------------------------------------------")
	println("Crawled url :\t" + items.GetRequest().GetUrl() + "\n")
	println("Crawled result : ")
	for key, value := range items.GetAll() {
		println(key + "\t:\t" + value)
	}
}

type PipelineFile struct {
	pFile *os.File

	path string
}

func NewPipelineFile(path string) *PipelineFile {
	pFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic("File '" + path + "' in PipelineFile open failed.")
	}
	return &PipelineFile{path: path, pFile: pFile}
}

func (this *PipelineFile) Process(items *PageItems, t Task) {
	this.pFile.WriteString("----------------------------------------------------------------------------------------------\n")
	this.pFile.WriteString("Crawled url :\t" + items.GetRequest().GetUrl() + "\n")
	this.pFile.WriteString("Crawled result : \n")
	for key, value := range items.GetAll() {
		this.pFile.WriteString(key + "\t:\t" + value + "\n")
	}
}
