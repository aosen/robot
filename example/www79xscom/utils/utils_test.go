/*
Author: Aosen
Date: 2016-01-19
QQ: 316052486
Desc: 工具箱
*/
package utils

import (
	"log"
	"testing"
)

func TestDownloadImage(t *testing.T) {
	url := "http://www.79xs.com/DownFiles/Book/BookCover/20151002204500_9456.gif"
	log.Println(DownloadImage(url, "static/novel/"))
}
