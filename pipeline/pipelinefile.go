/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: Pipeline file
*/
package pipeline

import (
	"os"

	"github.com/aosen/robot"
)

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

func (self *PipelineFile) Process(items *robot.PageItems, t robot.Task) {
	self.file.WriteString("----------------------------------------------------------------------------------------------\n")
	self.file.WriteString("Crawled url :\t" + items.GetRequest().GetUrl() + "\n")
	self.file.WriteString("Crawled result : \n")
	for key, value := range items.GetAll() {
		self.file.WriteString(key + "\t:\t" + value + "\n")
	}
}
